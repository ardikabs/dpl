package errs

import (
	"errors"
	"fmt"
)

type unwrapper interface {
	Unwrap() []error
}

// joinError is a custom error designed to combine multiple errors.
// It is heavily inspired by how errors.Join works, but it is specifically tailored to display error messages (err.Error()) in a chained format
// such as "new error: cause of error".
type joinError struct {
	err error
}

func (e *joinError) Error() string {
	u, ok := e.err.(unwrapper)
	if !ok {
		return e.err.Error()
	}

	var lastError error
	for _, err := range u.Unwrap() {
		if lastError == nil {
			lastError = err
			continue
		}

		lastError = fmt.Errorf("%v: %w", err, lastError)
	}

	return lastError.Error()
}

func (e *joinError) Unwrap() []error {
	if u, ok := e.err.(unwrapper); ok {
		return u.Unwrap()
	}

	return nil
}

func join(errs ...error) error {
	return &joinError{errors.Join(errs...)}
}
