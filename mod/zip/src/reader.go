package zip

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"io/fs"
)

var _ storage.Reader = &Reader{}

type Reader struct {
	fs.File
	name string
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	if s, ok := r.File.(io.Seeker); ok {
		return s.Seek(offset, whence)
	}

	return 0, errors.New("unsupported")
}

func (r *Reader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: r.name}
}
