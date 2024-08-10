package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

var _ astral.Object = &ArchiveDescriptor{}

type ArchiveDescriptor struct {
	Format    string
	Entries   uint32
	TotalSize uint64
}

func (d *ArchiveDescriptor) ObjectType() string {
	return "astrald.mod.archives.archive_descriptor"
}

func (d *ArchiveDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	err = cslq.Encode(w, "[c]clq", d.Format, d.Entries, d.TotalSize)
	if err == nil {
		n = int64(len(d.Format) + 1 + 4 + 8)
	}
	return
}

func (d *ArchiveDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	err = cslq.Decode(r, "[c]clq", &d.Format, &d.Entries, &d.TotalSize)
	if err == nil {
		n = int64(len(d.Format) + 1 + 4 + 8)
	}
	return
}

func (d *ArchiveDescriptor) String() string {
	return d.ObjectType()
}
