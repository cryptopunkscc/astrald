package astral

import (
	"fmt"
)

// ErrRejected - the query was rejected by the target
//var ErrRejected = errors.New("query rejected")

type ErrRejected struct {
	Code uint8
}

var _ error = &ErrRejected{}

func NewErrRejected(code uint8) *ErrRejected {
	if code == 0 {
		code = DefaultRejectCode
	}
	return &ErrRejected{Code: code}
}

func (e *ErrRejected) Error() string {
	return fmt.Sprintf("query rejected (%d)", e.Code)
}

func (e *ErrRejected) Is(other error) bool {
	_, ok := other.(*ErrRejected)
	return ok
}
