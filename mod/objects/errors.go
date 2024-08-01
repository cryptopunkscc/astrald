package objects

import "errors"

var (
	ErrNotFound           = errors.New("object not found")
	ErrObjectTooLarge     = errors.New("object too large")
	ErrSeekUnavailable    = errors.New("seek unavailable")
	ErrInvalidOffset      = errors.New("invalid offset")
	ErrStorageUnavailable = errors.New("storage unavailable")
	ErrAlreadyExists      = errors.New("already exists")
	ErrNoSpaceLeft        = errors.New("no space left on device")
	ErrClosedPipe         = errors.New("pipe closed")
	ErrAccessDenied       = errors.New("access denied")
	ErrHashMismatch       = errors.New("hash mismatch (data corrupted?)")
	ErrPushRejected       = errors.New("push rejected")
)
