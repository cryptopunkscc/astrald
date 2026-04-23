package nodes

import "errors"

var (
	ErrInvalidEndpointFormat = errors.New("invalid endpoint format")
	ErrEndpointParse         = errors.New("endpoint parse failed")
	ErrIdentityResolve       = errors.New("identity resolve failed")
	ErrEndpointResolve       = errors.New("endpoint resolve failed")
	ErrExcessStream          = errors.New("excess stream")
	ErrStreamNotProduced     = errors.New("stream not produced")
	ErrInvalidSessionState   = errors.New("invalid session state")
	ErrMigrationNotSupported = errors.New("migration not supported")
	ErrSessionNotFound       = errors.New("session not found")
)
