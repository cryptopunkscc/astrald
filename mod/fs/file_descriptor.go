package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"time"
)

var _ astral.Object = &FileDescriptor{}

type FileDescriptor struct {
	Path    string
	ModTime time.Time
}

func (d *FileDescriptor) ObjectType() string {
	return "astrald.mod.fs.file_descriptor"
}

func (d *FileDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	c := streams.NewWriteCounter(w)
	err = cslq.Encode(c, "[s]c v", d.Path, d.ModTime)
	n = c.Total()
	return
}

func (d *FileDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	c := streams.NewReadCounter(r)
	err = cslq.Decode(c, "[s]c v", &d.Path, &d.ModTime)
	n = c.Total()
	return
}

func (d *FileDescriptor) String() string {
	return d.Path
}
