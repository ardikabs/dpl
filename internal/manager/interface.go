package manager

import (
	"context"

	"github.com/ardikabs/dpl/internal/types"
)

type Manager interface {
	GetRelease(ctx context.Context, req *ReleaseRequest, opts ...Option) (*types.Release, error)
	SyncRelease(ctx context.Context, req *ReleaseRequest, opts ...Option) error
}
