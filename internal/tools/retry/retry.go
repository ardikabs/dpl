package retry

import (
	"context"
	"errors"
	"time"

	"github.com/ardikabs/dpl/internal/errs"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	ErrTimeout = errors.New("timed out waiting for the function to complete")
)

type RetriableFn func(error) bool

func OnError(ctx context.Context, retriable RetriableFn, fn func(ctx context.Context) error, opts ...RetryOption) error {
	options := &RetryOptions{
		Interval: time.Second,
		Timeout:  time.Second * 5,
	}

	for _, opt := range opts {
		opt(options)
	}

	var lastErr error

	log := options.Logger.WithName("retry.OnError")

	conditional := func(ctx context.Context) (done bool, err error) {
		err = fn(ctx)

		switch {
		case err == nil:
			return true, nil
		case retriable(err):
			log.V(2).Info("error caught, retrying ...", "err", err)

			lastErr = err
			return false, nil
		default:
			return false, err
		}
	}

	err := wait.PollUntilContextTimeout(ctx, options.Interval, options.Timeout, false, conditional)
	if errs.IsAny(err, context.DeadlineExceeded) {
		return errs.Wrap(lastErr, ErrTimeout)
	}

	return err
}
