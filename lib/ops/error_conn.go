package ops

import "io"

// ErrorConn is an io.ReadWriteCloser that returns an error on all operations.
type ErrorConn struct {
	Err error
}

var _ io.ReadWriteCloser = &ErrorConn{}

func (e ErrorConn) Read(p []byte) (n int, err error) {
	return 0, e.Err
}

func (e ErrorConn) Write(p []byte) (n int, err error) {
	return 0, e.Err
}

func (e ErrorConn) Close() error {
	return nil
}
