package nearby

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StatusMessage struct {
	Attachments *astral.Bundle
}

// astral

func (p *StatusMessage) ObjectType() string { return "mod.nearby.status_message" }

func (p StatusMessage) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&p).WriteTo(w)
}

func (p *StatusMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(p).ReadFrom(r)
}

// ...

func init() {
	_ = astral.Add(&StatusMessage{})
}
