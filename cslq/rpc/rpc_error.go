package rpc

// RPCError is an error extended with an error code, so that it can be easily encoded/decoded in binary protocols.
type RPCError struct {
	code  int
	error error
}

func (err *RPCError) Error() string {
	if err.error == nil {
		return ""
	}
	return err.error.Error()
}

func (err *RPCError) ErrorCode() int {
	return err.code
}

func (err *RPCError) Unwrap() error {
	return err.error
}
