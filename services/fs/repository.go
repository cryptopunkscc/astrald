package fs

import "io"

type Repository interface {
	Reader(id ID) (Reader, error)
	Writer() (Writer, error)
}

type Reader interface {
	io.ReadCloser
	Size() (int64, error)
}

type Writer interface {
	io.Writer
	Finalize() (*ID, error)
}