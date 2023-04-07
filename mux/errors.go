package mux

import "errors"

var ErrInvalidPort = errors.New("invalid port")
var ErrFrameTooLarge = errors.New("frame too large")
var ErrPortInUse = errors.New("port in use")
var ErrPortNotInUse = errors.New("port not in use")
var ErrAllPortsUsed = errors.New("all ports used")
var ErrPortClosed = errors.New("port closed")
var ErrCloseUnsupported = errors.New("transport does not support closing")
