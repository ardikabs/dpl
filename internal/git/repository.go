package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/ardikabs/dpl/internal/errs"
	"github.com/ardikabs/dpl/internal/tools/retry"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

var (
	ErrPullFailed    = errors.New("failed to pull from remote repository")
	ErrPushFailed    = errors.New("failed to push to remote repository")
	ErrRevisionReset = errors.New("revision reset because of non-fast-forward update error")
)

type gitRepo interface {
	Worktree() (*git.Worktree, error)
	Push(o *git.PushOptions) error
	Head() (*plumbing.Reference, error)
	FetchContext(ctx context.Context, o *git.FetchOptions) error
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

	head, err := g.gitRepo.Head()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(nil)
	cmd := exec.Command("git", "pull", RemoteName, head.Name().Short())
	cmd.Dir = g.Root()
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	log.V(1).Info("pull updates from remote repository with 'git' command", "cmd", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w, %v", ErrPullFailed, err)
	}

	worktree, err := g.gitRepo.Worktree()
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	log.V(2).Info("pull updates successfully", "clean", status.IsClean(), "cmdOutput", buffer.String())
	return nil
}

func (g *GitRepository) Commit(ctx context.Context, opts ...CommitOption) error {
	o := NewDefaultCommitOptions()
	for _, opt := range opts {
		opt(o)
	}

	log := o.Logger.WithName("repository.Commit")

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
		log.V(2).Info("skip commit, as because worktree is clean")
		return nil
	}

	log.V(2).Info("adding changes", "paths", o.Paths)
	for _, path := range o.Paths {
		if _, err := worktree.Add(path); err != nil {
			return err
		}
	}

	log.V(2).Info("commit changes", "message", o.Message)
	if _, err := worktree.Commit(o.Message, &git.CommitOptions{
		All:    true,
		Author: o.Committer,
	}); err != nil {
		return err
	}

	return nil
}

func (g *GitRepository) Push(ctx context.Context, opts ...PushOption) error {
	o := new(PushOptions)
	for _, opt := range opts {
		opt(o)
	}

	log := o.Logger.WithName("repository.Push")

	err := retry.OnError(ctx, func(err error) bool {
		if errs.IsAny(err, git.ErrNonFastForwardUpdate, ErrPullFailed, ErrPushFailed) {
			return true
		}

		return false
	}, func(ctx context.Context) error {
		// Ensure the git repository is up-to-date before push
		if err := g.Pull(ctx, WithPullLogger(log)); err != nil {
			return err
		}

		log.V(2).Info("push changes to remote repository")
		if err := g.gitRepo.Push(&git.PushOptions{Auth: g.auth}); err != nil {
			if !errs.IsAny(err, git.NoErrAlreadyUpToDate, transport.ErrEmptyRemoteRepository) {
				return fmt.Errorf("%w, %v", ErrPushFailed, err)
			}
		}

		return nil

	}, retry.WithRetryIntervalSec(1), retry.WithRetryTimoutSec(15), retry.WithLogger(log))

	if err != nil {
		return err
	}

	log.Info("worktree is up-to-date")
	return nil
}
