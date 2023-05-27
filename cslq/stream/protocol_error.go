package stream

// ProtocolError is an error extended with an error code, so that it can be easily encoded/decoded in binary protocols.
type ProtocolError struct {
	code  int
	error error
}

func (err *ProtocolError) Error() string {
	if err.error == nil {
		return ""
	}
	return err.error.Error()
}

func (err *ProtocolError) ErrorCode() int {
	return err.code
}

func (err *ProtocolError) Unwrap() error {
	return err.error
}
