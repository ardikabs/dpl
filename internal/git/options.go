package git

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-logr/logr"
)

type defaultOptions struct {
	RequestID string
	Logger    logr.Logger
}

type CloneOptions struct {
	defaultOptions

	SingleBranch bool
	Reference    plumbing.ReferenceName
}

func NewDefaultCloneOptions() *CloneOptions {
	return &CloneOptions{
		SingleBranch: true,
		Reference:    plumbing.HEAD,
	}
}

type CloneOption func(*CloneOptions)

func WithCloneBranch(branch string) CloneOption {
	return func(o *CloneOptions) {
		o.Reference = plumbing.ReferenceName(branch)
	}
}

func WithCloneLogger(logger logr.Logger) CloneOption {
	return func(o *CloneOptions) {
		o.Logger = logger
	}
}

type CommitOptions struct {
	defaultOptions

	Paths     []string
	Message   string
	Committer *object.Signature

	IsResetOnPushError bool
}

type CommitOption func(*CommitOptions)

func NewDefaultCommitOptions() *CommitOptions {
	return &CommitOptions{
		Message: "dpl: update deployment manifest",
		Committer: &object.Signature{
			Name:  "Deployment Auto BOT",
			Email: "bot@ardikabs.com",
			When:  time.Now(),
		},
	}
}

func WithCommitter(user, email string) CommitOption {
	return func(o *CommitOptions) {
		o.Committer = &object.Signature{
			Name:  user,
			Email: email,
			When:  time.Now(),
		}
	}
}

func WithCommitMessage(message string) CommitOption {
	return func(o *CommitOptions) {
		o.Message = message
	}
}

func WithCommitPath(paths ...string) CommitOption {
	return func(o *CommitOptions) {
		if len(paths) > 0 {
			o.Paths = []string{"."}
			return
		}

		o.Paths = paths
	}
}

func WithCommitLogger(logger logr.Logger) CommitOption {
	return func(o *CommitOptions) {
		o.Logger = logger
	}
}

func WithCommitResetOnPushError() CommitOption {
	return func(o *CommitOptions) {
		o.IsResetOnPushError = true
	}
}

type PullOptions struct {
	defaultOptions
}

type PullOption func(*PullOptions)

func WithPullLogger(logger logr.Logger) PullOption {
	return func(o *PullOptions) {
		o.Logger = logger
	}
}

type PushOptions struct {
	defaultOptions
}

type PushOption func(*PushOptions)

func WithPushLogger(logger logr.Logger) PushOption {
	return func(o *PushOptions) {
		o.Logger = logger
	}
}
