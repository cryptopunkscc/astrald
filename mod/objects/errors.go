package objects

import "errors"

var (
	ErrNotFound        = errors.New("object not found")
	ErrTagNotSupported = errors.New("tag not supported")
	ErrObjectTooLarge  = errors.New("object too large")
	ErrOutOfBounds     = errors.New("offset or limit out of bounds")
	ErrNoSpaceLeft     = errors.New("no space left on device")
	ErrClosedPipe      = errors.New("pipe closed")
	ErrPushRejected    = errors.New("push rejected")
)
