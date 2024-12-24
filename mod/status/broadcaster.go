package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &Broadcaster{}

type Broadcaster struct {
	Identity    *astral.Identity
	Alias       astral.String8
	LastSeen    astral.Time
	Attachments *astral.Bundle
}

func (Broadcaster) ObjectType() string {
	return "astrald.mod.status.broadcaster"
}

func (b Broadcaster) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(b).WriteTo(w)
}

func (b *Broadcaster) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(b).ReadFrom(r)
}
