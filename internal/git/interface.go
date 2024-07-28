package git

import (
	"golang.org/x/net/context"
)

type Interface interface {
	Clone(ctx context.Context, url, dest string, opts ...CloneOption) (Repository, error)
}

type Repository interface {
	Root() string
	Pull(ctx context.Context, opts ...PullOption) error
	CommitAndPush(ctx context.Context, opts ...CommitOption) error
}
