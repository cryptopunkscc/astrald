package infra

import "errors"

var ErrUnsupportedNetwork = errors.New("unsupported network")
var ErrUnsupportedOperation = errors.New("network does not support this operation")
var ErrUnsupportedAddress = errors.New("network does not support this address")
var ErrDialTimeout = errors.New("dial timed out")
var ErrInvalidAddress = errors.New("invalid address")
var ErrConnectionRefused = errors.New("connection refused")
