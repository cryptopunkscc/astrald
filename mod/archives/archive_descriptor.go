package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &ArchiveDescriptor{}

type ArchiveDescriptor struct {
	Format    astral.String
	Entries   astral.Uint32
	TotalSize astral.Uint64
}

func (ArchiveDescriptor) ObjectType() string {
	return "mod.archives.archive_descriptor"
}

func (d ArchiveDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(d).WriteTo(w)
}

func (d *ArchiveDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(d).ReadFrom(r)
}

func (d *ArchiveDescriptor) String() string {
	return d.ObjectType()
}

func init() {
	astral.DefaultBlueprints.Add(&ArchiveDescriptor{})
}
