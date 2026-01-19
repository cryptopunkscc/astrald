package fs

import "errors"

var ErrNotAbsolute = errors.New("path not absolute")
var ErrInvalidPath = errors.New("invalid path")
