package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/serializer"
	"io"
)

type LocalRepository interface {
	ReadWriteRepository
	ObserveRepository
}

type RemoteRepository interface {
	ReadRepository
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
	serializer.ReadCloser
	Size() (int64, error)
}

type Writer interface {
	serializer.Writer
	Finalize() (*fid.ID, error)
}

type Observer interface {
	serializer.ReadCloser
}
