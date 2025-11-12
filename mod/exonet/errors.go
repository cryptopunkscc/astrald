package exonet

import "errors"

var ErrUnsupportedNetwork = errors.New("unsupported network")
var ErrUnsupportedOperation = errors.New("unsupported operation")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")
