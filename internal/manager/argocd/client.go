package argocd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/tools/retry"
	"github.com/ardikabs/dpl/internal/types"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	applicationv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	synccommon "github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/ptr"
)

var (
	ErrArgoCDApplicationNotExists = errors.New("application not exists")
	ErrGitRepoAndRevisionMismatch = errors.New("git repository and revision must be the same")
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

func (c *Client) listApplications(ctx context.Context, selector string) ([]applicationv1.Application, error) {
	con, appClient := c.argocdClient.NewApplicationClientOrDie()
	defer con.Close()

	appList, err := appClient.List(ctx, &applicationpkg.ApplicationQuery{
		Selector: ptr.To(selector),
	})
	if err != nil {
		return nil, err
	}

	return appList.Items, nil
}

func (c *Client) getApplication(ctx context.Context, appName string) (*applicationv1.Application, error) {
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

func (c *Client) ListReleases(ctx context.Context, req *manager.ListReleaseRequest, opts ...manager.Option) ([]*types.Release, error) {
	options := manager.NewDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	log := options.Logger.WithName("argocd.ListReleases").WithValues("selector", req.Selector)
	apps, err := c.listApplications(ctx, req.Selector)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("releases found", "releases", len(apps))
	return appsToReleases(req, apps)
}

func (c *Client) SyncReleases(ctx context.Context, rels []*types.Release, opts ...manager.Option) error {
	options := manager.NewDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	log := options.Logger.WithName("argocd.SyncRelease")

	g, ctx := errgroup.WithContext(ctx)
	for _, rel := range rels {
		rel := rel

		g.Go(func() error {
			conn, appClient := c.argocdClient.NewApplicationClientOrDie()
			defer conn.Close()

			currentApp, err := c.getApplication(ctx, rel.ID)
			if err != nil {
				return err
			}

			if err := retry.OnError(ctx, func(err error) bool {
				if errors.Is(err, ErrAnotherSyncInProgress) {
					return true
				}
				return false
			}, func(ctx context.Context) error {
				if _, err := appClient.Sync(ctx, &applicationpkg.ApplicationSyncRequest{Name: ptr.To(rel.ID)}); err != nil {
					status, ok := status.FromError(err)
					if !ok {
						return err
					}

					if status.Code() == codes.FailedPrecondition ||
						strings.ToLower(status.Message()) == ErrAnotherSyncInProgress.Error() {

						log.V(1).Info("another sync operation is in progress", "app", rel.ID)
						return ErrAnotherSyncInProgress
					}
					return err
				}

				return nil
			},
				retry.WithRetryIntervalSec(1),
				retry.WithRetryTimoutSec(int(options.TimeoutSec)),
				retry.WithLogger(log),
				retry.WithErrOnTimeout(ErrSyncTimeout),
			); err != nil {
				return err
			}

			log = log.WithValues("app", currentApp.Name)
			log.Info("application sync is triggered")

			if err := c.watch(ctx, currentApp, watchOnSync,
				manager.WithTimeoutSec(options.TimeoutSec),
				manager.WithLogger(log)); err != nil {
				return err
			}

			return nil
		})
	}

	return g.Wait()
}

func watchOnSync(log logr.Logger, app applicationv1.Application) (bool, error) {
	good, err := checkAppStatus(log, app)
	if err != nil {
		return false, nil
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
			return false, fmt.Errorf("sync failed. reason: %s", state.Message)
		}
	}

	return good, nil
}
