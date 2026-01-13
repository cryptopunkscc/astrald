package fs

import "errors"

var (
	ErrFileModified = errors.New("file modified")
	ErrInvalidPath  = errors.New("invalid path")
	ErrNotIndexed   = errors.New("not indexed")
	ErrNotFound     = errors.New("not found")
)
