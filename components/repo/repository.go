package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"io"
)

type Repository interface {
	ReadWriteRepository
	ObserveRepository
}

type ReadWriteRepository interface {
	ReadRepository
	WriteRepository
}

type ReadRepository interface {
	Reader(id fid.ID) (Reader, error)
	List() (io.ReadCloser, error)
}

type WriteRepository interface {
	Writer() (Writer, error)
}

type ObserveRepository interface {
	Observer() (Observer, error)
}

type Reader interface {
	io.ReadCloser
	Size() (int64, error)
}

type Writer interface {
	io.Writer
	Finalize() (*fid.ID, error)
}

type Observer interface {
	io.ReadCloser
}
