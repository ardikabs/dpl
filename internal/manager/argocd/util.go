package argocd

import (
	"fmt"

	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/types"
	applicationv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"

	"github.com/go-logr/logr"
)

func appsToReleases(req *manager.ListReleaseRequest, apps []applicationv1.Application) ([]*types.Release, error) {
	if len(apps) == 0 {
		return nil, ErrArgoCDApplicationNotExists
	}

	var gitRepoURL, gitRevision string
	releases := make([]*types.Release, 0, len(apps))

	for _, app := range apps {
		if gitRepoURL == "" && gitRevision == "" {
			gitRepoURL = app.Spec.Source.RepoURL
			gitRevision = app.Spec.Source.TargetRevision
		} else if gitRepoURL != app.Spec.Source.RepoURL || gitRevision != app.Spec.Source.TargetRevision {
			return nil, ErrGitRepoAndRevisionMismatch
		}

		releases = append(releases, &types.Release{
			ID:          app.Name,
			Name:        req.GetReleaseFrom(app.Labels),
			Environment: req.GetEnvironmentFrom(app.Labels),
			Cluster:     req.GetClusterFrom(app.Labels),
			GitURL:      app.Spec.Source.RepoURL,
			GitPath:     app.Spec.Source.Path,
			GitRevision: app.Spec.Source.TargetRevision,
		})
	}

	return releases, nil
}

func checkAppStatus(logger logr.Logger, app applicationv1.Application) (bool, error) {
	log := logger.WithValues("sync.status", app.Status.Sync.Status, "health.status", app.Status.Health.Status)

	log.Info("application reconciliation is on progress")

	synced, err := checkAppSyncStatus(log, app)
	if err != nil {
		return false, err
	}

	healthy, err := checkAppHealthStatus(log, app)
	if err != nil {
		return false, err
	}

	return synced && healthy, nil

}

func checkAppSyncStatus(log logr.Logger, app applicationv1.Application) (bool, error) {
	switch app.Status.Sync.Status {
	case applicationv1.SyncStatusCodeSynced:
		log.V(1).Info("application is synced")
		return true, nil
	case applicationv1.SyncStatusCodeUnknown:
		log.V(1).Info("application sync status is unknown")
		return false, ErrStatusSyncUnknown
	case applicationv1.SyncStatusCodeOutOfSync:
		log.V(1).Info("application is on out-of-sync state")
	}

	return false, nil
}

func checkAppHealthStatus(log logr.Logger, app applicationv1.Application) (bool, error) {
	switch app.Status.Health.Status {
	case health.HealthStatusHealthy:
		log.V(1).Info("application is healthy")
		return true, nil
	case health.HealthStatusUnknown:
		log.V(1).Info("application health status is unknown")
		return false, ErrStatusSyncUnknown
	case health.HealthStatusProgressing:
		log.V(1).Info("application health status is progressing")
	case health.HealthStatusSuspended:
		log.V(1).Info("application is suspended")
	case health.HealthStatusMissing:
		log.V(1).Info("application health status check is failed or missing")
	case health.HealthStatusDegraded:
		log.V(1).Info("application health degraded for some reason, please check")
		return false, fmt.Errorf("%w, %s", ErrStatusHealthDegraded, app.Status.Health.Message)
	}

	return false, nil
}
