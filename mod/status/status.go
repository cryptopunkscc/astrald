package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Status struct {
	Port        astral.Uint16
	Alias       astral.String8
	Attachments *astral.Bundle
}

// astral

func (p *Status) ObjectType() string { return "mod.status.status" }

func (p *Status) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(p).WriteTo(w)
}

func (p *Status) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(p).ReadFrom(r)
}

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&Status{})
}
