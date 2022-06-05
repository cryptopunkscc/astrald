package hub

import (
	"errors"
)

// ErrAlreadyRegistered - provided port is already taken
var ErrAlreadyRegistered = errors.New("port already registered")

// ErrRejected - connection request was rejected
var ErrRejected = errors.New("rejected")

// ErrPortNotFound - provided port has not been registered
var ErrPortNotFound = errors.New("port not found")

// ErrTimeout - connection request timed out
var ErrTimeout = errors.New("timeout")

// ErrQueueOverflow - request queue is full, request could not be processed
var ErrQueueOverflow = errors.New("queue overflow")
