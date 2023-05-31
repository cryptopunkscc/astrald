package services

import (
	"errors"
)

// ErrAlreadyRegistered - provided port is already taken
var ErrAlreadyRegistered = errors.New("service already registered")

// ErrRejected - connection request was rejected
var ErrRejected = errors.New("rejected")

// ErrQueryHandled - Reject/Accept called more than once
var ErrQueryHandled = errors.New("query already handled")

// ErrServiceNotFound - provided port has not been registered
var ErrServiceNotFound = errors.New("service not found")

// ErrTimeout - connection request timed out
var ErrTimeout = errors.New("timeout")

// ErrQueueOverflow - request queue is full, request could not be processed
var ErrQueueOverflow = errors.New("query queue overflow")
