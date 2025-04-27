package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type RepositoryInfo struct {
	ID    astral.String8
	Label astral.String8
	Free  astral.Uint64
}

var _ astral.Object = &RepositoryInfo{}

func (RepositoryInfo) ObjectType() string {
	return "mod.objects.repository_info"
}

func (info RepositoryInfo) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(info).WriteTo(w)
}

func (info *RepositoryInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(info).ReadFrom(r)
}
