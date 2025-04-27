package fs

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io/fs"
)

var _ fs.File = &File{}

type File struct {
	ID *object.ID
	objects.Reader
}

func (f File) Stat() (fs.FileInfo, error) {
	return FileInfo{File: &f}, nil
}
