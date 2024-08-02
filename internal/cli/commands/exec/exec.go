package exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ardikabs/dpl/internal/git"
	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/manager/argocd"
	"github.com/ardikabs/dpl/internal/renderer"
	"github.com/ardikabs/dpl/internal/types"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

type execInstance struct {
	Params   *parameters
	Git      git.Interface
	Manager  manager.Interface
	Renderer renderer.Interface
	Logger   logr.Logger
}

func newExecInstance(log logr.Logger, params *parameters) (*execInstance, error) {
	g, err := git.New(params.GetGitSecret())
	if err != nil {
		return nil, err
	}

	argo, err := argocd.NewClient(types.ArgoConfig{
		Host:    params.ArgoCDHost,
		GRPCWeb: true,
		Secret: types.ArgoSecret{
			Token: params.ArgoCDAuthToken,
		},
	})
	if err != nil {
		return nil, err
	}

	return &execInstance{
		Git:      g,
		Manager:  argo,
		Renderer: renderer.New(params.Profile),
		Logger:   log,
		Params:   params,
	}, nil
}

func (ins *execInstance) Exec(ctx context.Context) error {
	imageDefinition := ins.Params.GetImageDefinition()

	reqID := uuid.New().String()
	log := ins.Logger.
		WithName("exec").
		WithValues(
			"release", ins.Params.ReleaseName,
			"environment", ins.Params.Environment,
			"image", imageDefinition.String(),
			"requestID", reqID,
		)

	req, err := manager.NewListReleaseRequestBuilder().
		SetReleaseSelector(ins.Params.SelectorForRelease, ins.Params.ReleaseName).
		SetEnvironmentSelector(ins.Params.SelectorForEnvironment, ins.Params.Environment).
		SetClusterSelector(ins.Params.SelectorForCluster, ins.Params.Cluster).
		Build()
	if err != nil {
		return err
	}

	releases, err := ins.Manager.ListReleases(ctx, req, manager.WithLogger(log))
	if err != nil {
		return err
	}

	gitURL := releases.GetGitURL()
	gitRevision := releases.GetGitRevision()

	log = log.WithValues("gitURL", gitURL, "gitRevision", gitRevision)

	workspace, err := os.MkdirTemp("/tmp", "dpl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)

	repo, err := ins.Git.Clone(ctx, gitURL, workspace, git.WithCloneBranch(gitRevision), git.WithCloneLogger(log))
	if err != nil {
		return err
	}

	for _, rel := range releases {
		log := log.WithValues("id", rel.ID, "cluster", rel.Cluster, "gitPath", rel.GitPath)
		rendererOpts := []renderer.RenderOption{
			renderer.WithLogger(log),
		}

		if ins.Params.IsTriggerRestart {
			rendererOpts = append(rendererOpts, renderer.WithExternalAnnotations(map[string]string{
				"dpl/restartedAt": time.Now().Format(time.RFC3339),
			}))
		}

		workdir := filepath.Join(repo.Root(), rel.GitPath)
		if err := ins.Renderer.Render(workdir, ins.Params.ReleaseName, &renderer.KustomizeParams{
			KustomizationRef:   ins.Params.KustomizationFileRef,
			ImageReferenceName: ins.Params.KustomizationImageRef,
			ImageName:          imageDefinition.Name,
			ImageTag:           imageDefinition.Tag,
		}, rendererOpts...); err != nil {
			return err
		}
	}

	if err := repo.Commit(ctx,
		git.WithCommitMessage(fmt.Sprintf("dpl(%s): update deployment manifest", reqID)),
		git.WithCommitter("kadabra-bot", "me@ardikabs"),
		git.WithCommitPath("."),
		git.WithCommitLogger(log),
	); err != nil {
		return err
	}

	if err := repo.Push(ctx, git.WithPushLogger(log)); err != nil {
		return err
	}

	if err := ins.Manager.SyncReleases(ctx, releases, manager.WithLogger(log)); err != nil {
		return err
	}

	log.Info("deployment executed successfully")
	return nil
}
