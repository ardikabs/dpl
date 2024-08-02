package errs

import (
	"errors"
	"fmt"
)

// IsAny is an extended version of errors.Is,
// that could check multiple errors at once
func IsAny(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

// Wrap a cause of error with an additional error
func Wrap(cause error, err error) error {
	return join(cause, err)
}

// Wrapf returns an error annotating cause of actual err with a custom message, then wrapped it
func Wrapf(cause error, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s; %w", msg, cause)
}
