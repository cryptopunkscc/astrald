package nodes

import "errors"

var (
	ErrInvalidEndpointFormat = errors.New("invalid endpoint format")
	ErrEndpointParse         = errors.New("endpoint parse failed")
	ErrIdentityResolve       = errors.New("identity resolve failed")
	ErrEndpointResolve       = errors.New("endpoint resolve failed")
	ErrNoEndpointReached     = errors.New("no endpoints reached")
	ErrExcessStream          = errors.New("excess stream")
	ErrStreamNotProduced     = errors.New("stream not produced")
	ErrInvalidMigrationState = errors.New("invalid migration state")
	ErrInvalidSessionState   = errors.New("invalid session state")
	ErrStreamNotFound        = errors.New("stream not found")
)
