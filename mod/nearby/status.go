package nearby

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Status struct {
	Identity    *astral.Identity
	Attachments *astral.Bundle
}

func (Status) ObjectType() string { return "mod.nearby.status" }

func (s Status) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *Status) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(s).ReadFrom(r)
}

func init() {
	_ = astral.Add(&Status{})
}
