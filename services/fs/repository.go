package fs

import "io"

type Repository interface {
	Reader(id ID) (Reader, error)
	Writer() (Writer, error)
}

type Reader interface {
	io.ReadCloser
}

type Writer interface {
	io.Writer
	Finalize() (*ID, error)
}