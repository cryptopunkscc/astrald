package sets

import (
	"errors"
	"fmt"
)

var ErrSetNotFound = errors.New("set not found")
var ErrMemberNotFound = errors.New("set member not found")
var ErrInvalidSetType = errors.New("invalid set type")

type ErrDatabaseError struct {
	err error
}

func DatabaseError(e error) *ErrDatabaseError {
	return &ErrDatabaseError{err: e}
}

func (e *ErrDatabaseError) Error() string {
	return fmt.Sprintf("database error: %s", e.err.Error())
}

func (e *ErrDatabaseError) Unwrap() error {
	return e.err
}

func (e *ErrDatabaseError) Is(other error) bool {
	_, same := other.(*ErrDatabaseError)
	return same
}
