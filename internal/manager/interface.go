package manager

import (
	"context"

	"github.com/ardikabs/dpl/internal/types"
)

type Manager interface {
	ListReleases(ctx context.Context, req *ListReleaseRequest, opts ...Option) ([]*types.Release, error)
	SyncReleases(ctx context.Context, rel []*types.Release, opts ...Option) error
}
