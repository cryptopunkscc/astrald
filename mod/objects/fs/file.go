package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"io/fs"
)

var _ fs.File = &File{}

type File struct {
	ID *object.ID
	objects.Reader
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	if s, ok := f.Reader.(io.Seeker); ok {
		return s.Seek(offset, whence)
	}
	return -1, errors.ErrUnsupported
}

func (f File) Stat() (fs.FileInfo, error) {
	return FileInfo{File: &f}, nil
}
