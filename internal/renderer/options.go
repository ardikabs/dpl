package renderer

import (
	"io"

	"github.com/go-logr/logr"
)

type RenderOptions struct {
	CustomWriter io.Writer
	Logger       logr.Logger

	Namespace string
}

type RenderOption func(*RenderOptions)

func WithNamespace(namespace string) RenderOption {
	return func(opts *RenderOptions) {
		opts.Namespace = namespace
	}
}

func WithCustomWriter(w io.Writer) RenderOption {
	return func(opts *RenderOptions) {
		opts.CustomWriter = w
	}
}

func WithLogger(logger logr.Logger) RenderOption {
	return func(opts *RenderOptions) {
		opts.Logger = logger
	}
}
