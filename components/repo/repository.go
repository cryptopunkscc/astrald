package repo

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/sio"
	"io"
)

type LocalRepository interface {
	ReadWriteMapRepository
	ObserveRepository
}

type RemoteRepository interface {
	ReadRepository
	ObserveRepository
}

type ReadWriteMapRepository interface {
	ReadWriteRepository
	MapperRepository
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

type MapperRepository interface {
	Map(path string) (*fid.ID, error)
}

type Reader interface {
	sio.ReadCloser
	Size() (int64, error)
}

type Writer interface {
	sio.Writer
	Finalize() (*fid.ID, error)
}

type Observer interface {
	sio.ReadCloser
}
