package mux

import "errors"

// ErrTooManyStreams - there are too many streams open.
var ErrTooManyStreams = errors.New("too many streams")

// ErrStreamNotFound - stream with the provided ID does not exist and cannot be closed or written to.
var ErrStreamNotFound = errors.New("stream not found")

// ErrPayloadTooBig - payload size exceeds the limit
var ErrPayloadTooBig = errors.New("payload too big")

// ErrReadError - error reading from the transport
var ErrReadError = errors.New("read error")
