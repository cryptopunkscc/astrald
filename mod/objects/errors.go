package objects

import "errors"

var (
	ErrNotFound       = errors.New("object not found")
	ErrObjectTooLarge = errors.New("object too large")
	ErrOutOfBounds    = errors.New("offset or limit out of bounds")
	ErrNoSpaceLeft    = errors.New("no space left on device")
	ErrClosedPipe     = errors.New("pipe closed")
	ErrPushRejected   = errors.New("push rejected")

	ErrNilSourceIdentifier   = errors.New("source identifier is nil")
	ErrInvalidSourceIdentity = errors.New("source identity is invalid")

	ErrAppRegistrationFromNetwork = errors.New("app discoverer registration cannot come from the network")
	ErrAppRegistrationSelf        = errors.New("node identity cannot register as an app discoverer")
)
