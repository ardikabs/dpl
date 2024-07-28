package argocd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardikabs/dpl/internal/manager"
	applicationv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
)

type appConditionFunc func(log logr.Logger, app applicationv1.Application) (bool, error)

func (c *Client) watch(ctx context.Context, app *applicationv1.Application, condition appConditionFunc, opts ...manager.Option) error {
	options := manager.NewDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(options.TimeoutSec)*time.Second)
	defer cancel()

	log := options.Logger.WithValues("operation", "sync/watch")

	var unknownRetryCount uint

	appCh := c.argocdClient.WatchApplicationWithRetry(ctx, app.Name, app.Spec.Source.TargetRevision)

	for {
		select {
		case ev, isOpen := <-appCh:
			if !isOpen {
				return ErrStatusSyncUnknown
			}

			app := ev.Application
			good, err := condition(log, app)
			if err != nil {
				if errors.Is(err, ErrStatusSyncUnknown) {
					if unknownRetryCount > uint(options.MaxRetryUnknownCount) {
						return err
					}

					// Trigger refresh
					if _, err := c.getApplicationWithName(ctx, app.Name); err != nil {
						return err
					}

					unknownRetryCount++
					continue
				}

				return err
			}

			if good {
				log.Info("watch completed", "status", app.Status.Sync.Status)
				return nil
			}

		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("watch timeout is exceeded")
			}

			return nil
		}
	}
}
