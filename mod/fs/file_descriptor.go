package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &FileDescriptor{}

type FileDescriptor struct {
	Path    astral.String16
	ModTime astral.Time
}

func (d *FileDescriptor) ObjectType() string {
	return "astrald.mod.fs.file_descriptor"
}

func (d *FileDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(d).WriteTo(w)
}

func (d *FileDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(d).ReadFrom(r)
}

func (d *FileDescriptor) String() string {
	return d.Path.String()
}

func init() {
	astral.DefaultBlueprints.Add(&FileDescriptor{})
}
