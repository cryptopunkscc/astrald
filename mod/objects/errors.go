package objects

import "errors"

var ErrNotFound = errors.New("object not found")
var ErrSeekUnavailable = errors.New("seek unavailable")
var ErrInvalidOffset = errors.New("invalid offset")
var ErrStorageUnavailable = errors.New("storage unavailable")
var ErrNoVirtual = errors.New("virtual source excluded")
var ErrAlreadyExists = errors.New("already exists")
var ErrNoSpaceLeft = errors.New("no space left on device")
var ErrClosedPipe = errors.New("pipe closed")
var ErrZoneExcluded = errors.New("zone excluded")
var ErrAccessDenied = errors.New("access denied")
