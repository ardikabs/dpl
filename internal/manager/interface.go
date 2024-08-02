package manager

import (
	"context"

	"github.com/ardikabs/dpl/internal/types"
)

type Interface interface {
	ListReleases(ctx context.Context, req *ListReleaseRequest, opts ...Option) (types.ListReleases, error)
	SyncRelease(ctx context.Context, rel *types.Release, opts ...Option) error
	SyncReleases(ctx context.Context, rels types.ListReleases, opts ...Option) error
}
