package fs

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"io"
)

type Repository interface {
	Reader(id fid.ID) (Reader, error)
	Writer() (Writer, error)
}

type Reader interface {
	io.ReadCloser
	Size() (int64, error)
}

type Writer interface {
	io.Writer
	Finalize() (*fid.ID, error)
}
