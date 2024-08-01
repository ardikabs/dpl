package git

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ardikabs/dpl/internal/errs"
	"github.com/ardikabs/dpl/internal/tools/cmdutils"
	"github.com/ardikabs/dpl/internal/tools/retry"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-logr/logr"
)

var (
	ErrPullFailed = errors.New("failed to pull from remote repository")
	ErrPushFailed = errors.New("failed to push to remote repository")
)

type gitRepo interface {
	Head() (*plumbing.Reference, error)
	Worktree() (*git.Worktree, error)
	Push(o *git.PushOptions) error
	Config() (*config.Config, error)
	SetConfig(cfg *config.Config) error
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
		return "UNDEFINED"
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

	if err := worktree.PullContext(ctx, &git.PullOptions{
		Auth:  g.auth,
		Force: true,
	}); err != nil {
		if !errs.IsAny(err, git.NoErrAlreadyUpToDate, transport.ErrEmptyRemoteRepository) {
			return errs.Wrap(err, ErrPullFailed)
		}
	}

	log.V(2).Info("pull updates successfully")
	return nil
}

func (g *GitRepository) rawPull(ctx context.Context, log logr.Logger) error {
	log = log.WithName("repository.rawPull")

	if err := g.setGitRepoConfig(); err != nil {
		return err
	}

	head, err := g.gitRepo.Head()
	if err != nil {
		return err
	}

	opts := []cmdutils.Option{
		cmdutils.WithWorkdir(g.Root()),
		cmdutils.WithArgs("pull", RemoteName, head.Name().Short()),
		cmdutils.WithLogger(log),
	}

	if err := cmdutils.Exec(ctx, "git", opts...); err != nil {
		return errs.Wrap(err, ErrPullFailed)
	}

	log.V(1).Info("pull updates successfully")
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
		// for some reason like non-fast-forward update error, we use git raw command instead to pull
		// the reason because go-git doesn't support rebase mechanism,
		// hence non-fast-forward update become problematic
		if err := g.rawPull(ctx, log); err != nil {
			return err
		}

		log.V(2).Info("push changes to remote repository")
		if err := g.gitRepo.Push(&git.PushOptions{Auth: g.auth}); err != nil {
			if !errs.IsAny(err, git.NoErrAlreadyUpToDate, transport.ErrEmptyRemoteRepository) {
				return errs.Wrap(err, ErrPushFailed)
			}
		}

		return nil

	}, retry.WithRetryIntervalSec(1), retry.WithRetryTimoutSec(15), retry.WithLogger(log))

	if err != nil {
		if errs.IsAny(err, retry.ErrTimeout) {
			log.V(1).Info("push operation is timed out")
		}

		return err
	}

	log.Info("worktree is up-to-date")
	return nil
}

func (g *GitRepository) setGitRepoConfig() error {
	cfg, err := g.gitRepo.Config()
	if err != nil {
		return err
	}

	for _, remote := range cfg.Remotes {
		if remote.Name != RemoteName && len(remote.URLs) == 0 {
			continue
		}
		if httpAuth, ok := g.auth.(*http.BasicAuth); ok {
			url := remote.URLs[0]
			urlParts := strings.Split(url, "://")
			remote.URLs = []string{fmt.Sprintf("%s://%s:%s@%s", urlParts[0], httpAuth.Username, httpAuth.Password, urlParts[1])}
		}

	}

	for _, branch := range cfg.Branches {
		branch.Rebase = "true"
	}

	return g.gitRepo.SetConfig(cfg)
}
