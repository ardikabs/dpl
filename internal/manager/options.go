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

func NewDefaultOptions() *Options {
	return &Options{
		TimeoutSec:           DefaultTimeout,
		MaxRetryUnknownCount: 5,
	}
}

type Option func(*Options)

func WithTimeoutSec(timeoutSec uint) Option {
	return func(opts *Options) {
		opts.TimeoutSec = timeoutSec
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
