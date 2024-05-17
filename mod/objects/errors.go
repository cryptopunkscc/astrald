package objects

import "errors"

var ErrNotFound = errors.New("object not found")
var ErrObjectTooLarge = errors.New("object too large")
var ErrSeekUnavailable = errors.New("seek unavailable")
var ErrInvalidOffset = errors.New("invalid offset")
var ErrStorageUnavailable = errors.New("storage unavailable")
var ErrAlreadyExists = errors.New("already exists")
var ErrNoSpaceLeft = errors.New("no space left on device")
var ErrClosedPipe = errors.New("pipe closed")
var ErrAccessDenied = errors.New("access denied")
