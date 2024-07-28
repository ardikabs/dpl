package argocd

import (
	"fmt"

	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/types"
	applicationv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"

	"github.com/go-logr/logr"
)

func appToRelease(req *manager.ReleaseRequest, app *applicationv1.Application) *types.Release {
	return &types.Release{
		ID:          app.Name,
		Name:        req.GetReleaseFrom(app.Labels),
		Environment: req.GetEnvironmentFrom(app.Labels),
		Cluster:     req.GetClusterFrom(app.Labels),
		GitURL:      app.Spec.Source.RepoURL,
		GitPath:     app.Spec.Source.Path,
		GitRevision: app.Spec.Source.TargetRevision,
	}
}

func checkAppSyncStatus(logger logr.Logger, app applicationv1.Application) (bool, error) {
	log := logger.WithValues("sync.status", app.Status.Sync.Status)

	switch app.Status.Sync.Status {
	case applicationv1.SyncStatusCodeSynced:
		log.Info("application is synced")
		return true, nil
	case applicationv1.SyncStatusCodeUnknown:
		log.Info("application sync status is unknown")
		return false, ErrStatusSyncUnknown
	case applicationv1.SyncStatusCodeOutOfSync:
		log.Info("application is on out-of-sync state")
	}

	return false, nil
}

func checkAppHealthStatus(logger logr.Logger, app applicationv1.Application) (bool, error) {
	log := logger.WithValues("health.status", app.Status.Health.Status)

	switch app.Status.Health.Status {
	case health.HealthStatusHealthy:
		log.Info("application is healthy")
		return true, nil
	case health.HealthStatusUnknown:
		log.Info("application health status is unknown")
		return false, ErrStatusSyncUnknown
	case health.HealthStatusProgressing:
		log.Info("application health status is progressing")
	case health.HealthStatusSuspended:
		log.Info("application is suspended")
	case health.HealthStatusMissing:
		log.Info("application health status check is failed or missing")
	case health.HealthStatusDegraded:
		log.Info("application health degraded for some reason, please check")
		return false, fmt.Errorf("%w, %s", ErrStatusHealthDegraded, app.Status.Health.Message)
	}

	return false, nil
}
