package storage

import "errors"

var ErrNotFound = errors.New("not found")
var ErrSeekUnavailable = errors.New("seek unavailable")
var ErrInvalidOffset = errors.New("invalid offset")
var ErrStorageUnavailable = errors.New("storage unavailable")
var ErrNoVirtual = errors.New("virtual source excluded")
var ErrAlreadyExists = errors.New("already exists")
