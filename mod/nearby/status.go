package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Status struct {
	Identity    *astral.Identity
	Alias       astral.String8
	Attachments *astral.Bundle
}

func (Status) ObjectType() string { return "mod.nearby.status" }

func (s Status) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(s).WriteTo(w)
}

func (s Status) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(s).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&Status{})
}
