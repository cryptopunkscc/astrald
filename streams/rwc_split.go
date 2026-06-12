package streams

import "io"

// ReadWriteCloseSplit composes independent Reader, Writer, and Closer into a single ReadWriteCloser.
type ReadWriteCloseSplit struct {
	io.Reader
	io.Writer
	io.Closer
}
