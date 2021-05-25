package net

import "errors"

// ErrAlreadyRegistered - a different dial is already registered for the network
var ErrAlreadyRegistered = errors.New("network already registered")

// ErrInvalidNetworkName - provided network name is invalid
var ErrInvalidNetworkName = errors.New("invalid network name")

// ErrUnsupportedNetwork - requested network is not supported
var ErrUnsupportedNetwork = errors.New("unsupported network")

// ErrHostUnreachable - requested host could not be reached
var ErrHostUnreachable = errors.New("host unreachable")

// ErrUnsupported - requested feature is not available on this network
var ErrUnsupported = errors.New("unsupported feature")
