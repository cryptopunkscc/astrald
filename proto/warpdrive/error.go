package warpdrive

import (
	"fmt"
)

func Error(cause error, message ...interface{}) error {
	s := fmt.Sprintln(message...)
	s = s[:len(s)-1]
	return wrappedError{
		cause:   cause,
		message: s,
	}
}

type wrappedError struct {
	cause   error
	message string
}

func (e wrappedError) Error() string {
	s := fmt.Sprintln(e.message)
	if e.cause != nil {
		s = fmt.Sprint(s, " - ", e.cause)
	}
	return s
}

func (e wrappedError) String() string {
	return e.Error()
}

func (e wrappedError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}
	return e
}
