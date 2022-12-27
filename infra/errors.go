package infra

import "errors"

var ErrUnsupportedNetwork = errors.New("unsupported network")
var ErrUnsupportedAddress = errors.New("network does not support this address")
var ErrInvalidAddress = errors.New("invalid address")
var ErrConnectionRefused = errors.New("connection refused")
