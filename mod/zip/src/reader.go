package zip

import (
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io/fs"
)

var _ storage.DataReader = &Reader{}

type Reader struct {
	fs.File
	name string
}

func (r *Reader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: r.name}
}
