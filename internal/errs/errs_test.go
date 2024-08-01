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

func TestWrap(t *testing.T) {
	lowErr := errors.New("low level error")

	firstErr := errs.Wrapf(lowErr, "first error initiated with %s", "some reason")
	require.ErrorIs(t, firstErr, lowErr)
	require.Equal(t, "first error initiated with some reason; low level error", firstErr.Error())

	secondErr := errors.New("it is a second error")
	combinedErr := errs.Wrap(lowErr, secondErr)
	require.ErrorIs(t, combinedErr, lowErr)
	require.ErrorIs(t, combinedErr, secondErr)

	thirdErr := errors.New("it is a third error")
	anotherCombinedErr := errs.Wrap(thirdErr, nil)
	require.ErrorIs(t, anotherCombinedErr, thirdErr)
}
