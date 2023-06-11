package rpc

import "errors"

// ErrorSpace is a collection of ProtocolErrors which is used to encode/decode errors over a stream.
type ErrorSpace struct {
	err map[int]*RPCError
}

// NewError creates a new error and wraps it using WrapError.
func (space *ErrorSpace) NewError(code int, error string) (err *RPCError) {
	return space.WrapError(code, errors.New(error))
}

// WrapError wraps an error into a new error in the space. If another error with the same code exists,
// it will be overwritten. Code cannot be zero, since zero is reserved for no error (nil).
func (space *ErrorSpace) WrapError(code int, error error) (err *RPCError) {
	if code == 0 {
		panic("code cannot be zero")
	}
	if space.err == nil {
		space.err = map[int]*RPCError{}
	}
	err = &RPCError{
		code:  code,
		error: error,
	}
	space.err[code] = err
	return
}

// ByCode returns the error assigned to the code (or nil if the code is zero) and true. If the code is invalid it
// returns nil and false.
func (space *ErrorSpace) ByCode(code int) (*RPCError, bool) {
	if code == 0 {
		return nil, true
	}
	if space.err == nil {
		return nil, false
	}
	if err, found := space.err[code]; found {
		return err, true
	}
	return nil, false
}
