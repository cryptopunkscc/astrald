package nodes

import "errors"

var (
	ErrInvalidEndpointFormat = errors.New("invalid endpoint format")
	ErrEndpointParse         = errors.New("endpoint parse failed")
	ErrIdentityResolve       = errors.New("identity resolve failed")
	ErrEndpointResolve       = errors.New("endpoint resolve failed")
	ErrNoUsableEndpoints     = errors.New("no endpoints")
	ErrNoEndpointReached     = errors.New("no endpoints reached")
	ErrExcessStream          = errors.New("excess stream")
)
