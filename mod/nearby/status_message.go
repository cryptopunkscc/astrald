package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type StatusMessage struct {
	Port        astral.Uint16
	Alias       astral.String8
	Attachments *astral.Bundle
}

// astral

func (p *StatusMessage) ObjectType() string { return "mod.nearby.status_message" }

func (p *StatusMessage) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(p).WriteTo(w)
}

func (p *StatusMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(p).ReadFrom(r)
}

// ...

func init() {
	_ = astral.Add(&StatusMessage{})
}
