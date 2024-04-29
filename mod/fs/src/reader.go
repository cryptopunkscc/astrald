package fs

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"io"
)

var _ objects.Reader = &Reader{}

type Reader struct {
	io.ReadSeekCloser
	name string
}

func (r *Reader) Info() *objects.ReaderInfo {
	return &objects.ReaderInfo{Name: r.name}
}
