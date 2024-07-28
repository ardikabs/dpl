package exec

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ardikabs/dpl/internal/git"
	"github.com/ardikabs/dpl/internal/manager"
	"github.com/ardikabs/dpl/internal/manager/argocd"
	"github.com/ardikabs/dpl/internal/renderer"
	"github.com/ardikabs/dpl/internal/types"
	"github.com/go-logr/logr"
)

type execInstance struct {
	Params   *parameters
	Git      git.Interface
	Manager  manager.Manager
	Renderer renderer.Renderer
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

	var rdr renderer.Renderer
	switch params.Profile {
	case "kustomize":
		rdr = new(renderer.Kustomize)
	}

	return &execInstance{
		Git:      g,
		Manager:  argo,
		Renderer: rdr,
		Logger:   log,
		Params:   params,
	}, nil
}

func (ins *execInstance) Exec(ctx context.Context) error {
	log := ins.Logger.WithName("instance.Start")

	reqOpts := &manager.ReleaseRequestBuilderOptions{
		SelectorKeyForRelease:     ins.Params.SelectorForRelease,
		SelectorKeyForEnvironment: ins.Params.SelectorForEnvironment,
		SelectorKeyForCluster:     ins.Params.SelectorForCluster,
	}
	reqBuilder := manager.NewReleaseRequestBuilderWithOptions(reqOpts).
		SetReleaseSelector(ins.Params.ReleaseName).
		SetEnvironmentSelector(ins.Params.Environment)

	if ins.Params.Cluster != "" {
		reqBuilder = reqBuilder.SetClusterSelector(ins.Params.Cluster)
	}

	relReq := reqBuilder.Build()
	rel, err := ins.Manager.GetRelease(ctx, relReq, manager.WithLogger(log))
	if err != nil {
		return err
	}

	imageDefinition := ins.Params.GetImageDefinition()

	log = log.WithValues(
		"releaseID", rel.ID,
		"release", rel.Name,
		"cluster", rel.Cluster,
		"gitURL", rel.GitURL,
		"gitRevision", rel.GitRevision,
		"gitPath", rel.GitPath,
		"image", imageDefinition.String(),
	)

	workspace, err := os.MkdirTemp("/tmp", "dpl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)

	repo, err := ins.Git.Clone(ctx, rel.GitURL, workspace, git.WithCloneBranch(rel.GitRevision), git.WithCloneLogger(log))
	if err != nil {
		return err
	}

	workdir := filepath.Join(repo.Root(), rel.GitPath)
	if err := ins.Renderer.Render(workdir, ins.Params.ReleaseName, &renderer.KustomizeParams{
		KustomizationRef:   ins.Params.KustomizationFileRef,
		ImageReferenceName: ins.Params.KustomizationImageRef,
		ImageName:          imageDefinition.Name,
		ImageTag:           imageDefinition.Tag,
	}, renderer.WithLogger(log)); err != nil {
		return err
	}

	if err := repo.CommitAndPush(ctx,
		git.WithCommitMessage("dpl: update deployment manifest"),
		git.WithCommitter("kadabra-bot", "me@ardikabs"),
		git.WithCommitPath(rel.GitPath),
		git.WithCommitLogger(log),
	); err != nil {
		return err
	}

	if err := ins.Manager.SyncRelease(ctx, relReq, manager.WithLogger(log)); err != nil {
		return err
	}

	return nil
}
