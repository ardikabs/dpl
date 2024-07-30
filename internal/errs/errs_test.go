package errs_test

import (
	"errors"
	"testing"

	"github.com/ardikabs/dpl/internal/errs"
	"github.com/stretchr/testify/require"
)

func TestIsAny(t *testing.T) {
	errA := errors.New("this is error A")
	errB := errors.New("this is error B")

	require.True(t, errs.IsAny(errA, errA, errB))
	require.True(t, errs.IsAny(errB, errA, errB))
	require.False(t, errs.IsAny(errors.New("new error"), errA, errB))
}
