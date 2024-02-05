package fs

import (
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

var _ storage.Reader = &Reader{}

type Reader struct {
	io.ReadSeekCloser
	name string
}

func (r *Reader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: r.name}
}
