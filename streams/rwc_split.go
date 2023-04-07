package streams

import "io"

type ReadWriteCloseSplit struct {
	io.Reader
	io.Writer
	io.Closer
}
