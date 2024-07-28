package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

var (
	ErrFailedToPush = errors.New("failed to push to remote repository")
)

type gitRepo interface {
	Worktree() (*git.Worktree, error)
	Push(o *git.PushOptions) error
	Head() (*plumbing.Reference, error)
}

type GitRepository struct {
	gitRepo gitRepo
	auth    transport.AuthMethod
}

func NewGitRepository(repo gitRepo, auth transport.AuthMethod) *GitRepository {
	return &GitRepository{
		gitRepo: repo,
		auth:    auth,
	}
}

func (g *GitRepository) Root() string {
	worktree, err := g.gitRepo.Worktree()
	if err != nil {
		return ""
	}

	return worktree.Filesystem.Root()
}

func (g *GitRepository) Pull(ctx context.Context, opts ...PullOption) error {
	o := new(PullOptions)
	for _, opt := range opts {
		opt(o)
	}

	log := o.Logger.WithName("repository.Pull")

	worktree, err := g.gitRepo.Worktree()
	if err != nil {
		return err
	}

	log.V(1).Info("pull updates from remote repository")
	if err := worktree.PullContext(ctx, &git.PullOptions{
		Auth:       g.auth,
		RemoteName: "origin",
		Force:      true,
	}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) || errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return err
		}
	}

	return nil
}

func (g *GitRepository) CommitAndPush(ctx context.Context, opts ...CommitOption) error {
	o := NewDefaultCommitOptions()
	for _, opt := range opts {
		opt(o)
	}

	log := o.Logger.WithName("repository.CommitAndPush")

	worktree, err := g.gitRepo.Worktree()
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	// Skip commit and push if worktree is clean
	if status.IsClean() {
		log.V(1).Info("skip commit and push, as because worktree is clean")
		return nil
	}

	log.V(1).Info("adding changes", "paths", o.Paths)
	for _, path := range o.Paths {
		if _, err := worktree.Add(path); err != nil {
			return err
		}
	}

	log.V(1).Info("commit changes", "message", o.Message)
	if _, err := worktree.Commit(o.Message, &git.CommitOptions{
		All:    true,
		Author: o.Committer,
	}); err != nil {
		return err
	}

	// Ensure the git repository is up-to-date before push
	if err := g.Pull(ctx); err != nil {
		return err
	}

	log.V(1).Info("push changes to remote repository")
	if err := g.gitRepo.Push(&git.PushOptions{Auth: g.auth}); err != nil {
		return fmt.Errorf("%w, %v", ErrFailedToPush, err)
	}

	return nil
}
