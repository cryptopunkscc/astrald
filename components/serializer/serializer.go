package serializer

import "io"

type ReadWriteCloser struct {
	io.Closer
	*Reader
	*Writer
}
