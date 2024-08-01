package errs

import (
	"errors"
	"fmt"
)

func IsAny(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

func Wrap(cause error, err error) error {
	return join(cause, err)
}

func Wrapf(cause error, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s; %w", msg, cause)
}
