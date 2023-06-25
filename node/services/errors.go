package services

import (
	"errors"
)

// ErrAlreadyRegistered - provided port is already taken
var ErrAlreadyRegistered = errors.New("service already registered")

// ErrServiceNotFound - provided port has not been registered
var ErrServiceNotFound = errors.New("service not found")

// ErrTimeout - connection request timed out
var ErrTimeout = errors.New("timeout")
