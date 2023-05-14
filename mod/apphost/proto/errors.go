package proto

import (
	"errors"
)

const ErrorCSLQ = "c"

const (
	Success = 0x00
)

// protocol errors
var (
	ErrRejected          = makeError(0x01, "rejected")
	ErrFailed            = makeError(0x02, "failed")
	ErrTimeout           = makeError(0x03, "timeout")
	ErrAlreadyRegistered = makeError(0x04, "port already registered")
	ErrUnauthorized      = makeError(0x10, "unauthorized")
	ErrUnknownCommand    = makeError(0xfd, "unknown command")
	ErrUnexpected        = makeError(0xff, "unexpected error")
)

// local errors
var (
	ErrUnknown             = errors.New("unknown error")
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
	ErrInvalidIPCAddress   = errors.New("invalid ipc address")
)

// exit codes returned by the runtime
const (
	ExitCodeOK               = iota // no error
	ExitCodeInvalidArguments        // invalid command line arguments or arguments missing
	ExitCodeLoadError               // cannot load application
	ExitCodeAppError                // application ended with an error
)
