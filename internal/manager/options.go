package manager

import "github.com/go-logr/logr"

var (
	DefaultTimeout uint = 900 // 15 minutes
)

type Options struct {
	Logger               logr.Logger
	TimeoutSec           uint
	MaxRetryUnknownCount int
}

func NewDefaultOptions(opts ...Option) *Options {
	o := &Options{
		TimeoutSec:           DefaultTimeout,
		MaxRetryUnknownCount: 5,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(*Options)

func WithTimeoutSec(timeoutSec uint) Option {
	return func(opts *Options) {
		opts.TimeoutSec = timeoutSec
	}
}

func WithMaxRetryUnknownCount(maxRetryUnknownCount int) Option {
	return func(opts *Options) {
		opts.MaxRetryUnknownCount = maxRetryUnknownCount
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
