package retry

import (
	"time"

	"github.com/go-logr/logr"
)

type RetryOptions struct {
	Interval time.Duration
	Timeout  time.Duration
	Logger   logr.Logger
}

type RetryOption func(*RetryOptions)

func WithRetryIntervalSec(intervalSec int) RetryOption {
	return func(o *RetryOptions) {
		o.Interval = time.Duration(intervalSec) * time.Second
	}
}

func WithRetryTimoutSec(timeoutSec int) RetryOption {
	return func(o *RetryOptions) {
		o.Timeout = time.Duration(timeoutSec) * time.Second
	}
}

func WithLogger(logger logr.Logger) RetryOption {
	return func(o *RetryOptions) {
		o.Logger = logger
	}
}
