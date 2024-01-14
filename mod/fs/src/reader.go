package fs

import (
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

var _ storage.DataReader = &Reader{}

type Reader struct {
	io.ReadCloser
	name string
}

func (r *Reader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: r.name}
}
