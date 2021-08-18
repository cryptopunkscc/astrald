package serializer

import "io"

type readWriteCloser struct {
	io.Closer
	Reader
	Writer
}
