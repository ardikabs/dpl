package errs

import (
	"errors"
	"fmt"
)

type unwrapper interface {
	Unwrap() []error
}

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
