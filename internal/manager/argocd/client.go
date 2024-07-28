package argocd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/types"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	applicationv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	synccommon "github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

var (
	ErrArgoCDApplicationNotExists = errors.New("application not exists")
	ErrStatusSyncUnknown          = errors.New("sync status unknown")
	ErrStatusHealthDegraded       = errors.New("health status degraded")
	ErrAnotherSyncInProgress      = errors.New("another operation is already in progress")
	ErrSyncTimeout                = errors.New("sync timeout is exceeded")
)

type client interface {
	NewApplicationClientOrDie() (io.Closer, applicationpkg.ApplicationServiceClient)
	WatchApplicationWithRetry(ctx context.Context, appName string, revision string) chan *applicationv1.ApplicationWatchEvent
}

type Client struct {
	argocdClient client
}

func NewClient(cfg types.ArgoConfig) (*Client, error) {
	cl, err := apiclient.NewClient(&apiclient.ClientOptions{
		ServerAddr: cfg.Host,
		PlainText:  cfg.PlainText,
		Insecure:   cfg.Insecure,
		GRPCWeb:    cfg.GRPCWeb,
		AuthToken:  cfg.Secret.Token,
	})
	if err != nil {
		return nil, err
	}

	return &Client{argocdClient: cl}, nil
}

func (c *Client) getApplicationWithSelector(ctx context.Context, selector *string) (*applicationv1.Application, error) {
	con, appClient := c.argocdClient.NewApplicationClientOrDie()
	defer con.Close()

	appList, err := appClient.List(ctx, &applicationpkg.ApplicationQuery{
		Selector: selector,
		Refresh:  ptr.To("true"),
	})
	if err != nil {
		return nil, err
	}

	if len(appList.Items) > 0 {
		return &appList.Items[0], nil
	}

	return nil, ErrArgoCDApplicationNotExists
}

func (c *Client) getApplicationWithName(ctx context.Context, appName string) (*applicationv1.Application, error) {
	con, appClient := c.argocdClient.NewApplicationClientOrDie()
	defer con.Close()

	app, err := appClient.Get(ctx, &applicationpkg.ApplicationQuery{
		Name:    ptr.To(appName),
		Refresh: ptr.To("true"),
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (c *Client) GetRelease(ctx context.Context, req *manager.ReleaseRequest, opts ...manager.Option) (*types.Release, error) {
	options := manager.NewDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	log := options.Logger.WithName("argocd.GetRelease").
		WithValues("selector", ptr.Deref(req.Selector, "N/A"))

	app, err := c.getApplicationWithSelector(ctx, req.Selector)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("application found", "app", app.Name, "status", app.Status.Sync.Status)
	return appToRelease(req, app), nil
}

func (c *Client) SyncRelease(ctx context.Context, req *manager.ReleaseRequest, opts ...manager.Option) error {
	options := manager.NewDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	log := options.Logger.WithName("argocd.SyncRelease")

	rel, err := c.GetRelease(ctx, req, manager.WithLogger(log))
	if err != nil {
		return err
	}

	conn, appClient := c.argocdClient.NewApplicationClientOrDie()
	defer conn.Close()

	var currentApp *applicationv1.Application
	if err := wait.PollUntilContextTimeout(ctx, time.Second, time.Duration(options.TimeoutSec)*time.Second, false, func(ctx context.Context) (bool, error) {
		var err error
		currentApp, err = appClient.Sync(ctx, &applicationpkg.ApplicationSyncRequest{
			Name: ptr.To(rel.ID),
		})
		if err != nil {
			status, ok := status.FromError(err)
			if !ok {
				return false, err
			}

			if status.Code() == codes.FailedPrecondition ||
				strings.ToLower(status.Message()) == ErrAnotherSyncInProgress.Error() {

				log.Info(ErrAnotherSyncInProgress.Error())
				return false, nil
			}

			return false, err
		}

		return true, nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ErrSyncTimeout
		}

		return err
	}

	log = log.WithValues("app", currentApp.Name)
	log.Info("application sync is triggered", "status", currentApp.Status.Sync.Status)

	if err := c.watch(ctx, currentApp, watchOnSync,
		manager.WithTimeoutSec(options.TimeoutSec),
		manager.WithLogger(log)); err != nil {
		return err
	}

	return nil
}

func watchOnSync(log logr.Logger, app applicationv1.Application) (bool, error) {
	sync, err := checkAppSyncStatus(log, app)
	if err != nil {
		return false, err
	}

	health, err := checkAppHealthStatus(log, app)
	if err != nil {
		return false, err
	}

	state := app.Status.OperationState
	if state != nil {
		switch state.Phase {
		case synccommon.OperationError, synccommon.OperationFailed:
			for _, resource := range state.SyncResult.Resources {
				log.Info(resource.Message,
					"group", resource.Group,
					"kind", resource.Kind,
					"name", resource.Name,
					"namespace", resource.Namespace,
					"version", resource.Version,
					"status", resource.Status,
				)
			}
		}

		return false, fmt.Errorf("sync failed. reason: %s", state.Message)
	}

	return sync && health, nil
}
