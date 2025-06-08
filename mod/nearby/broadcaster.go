package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Broadcaster struct {
	Identity    *astral.Identity
	Alias       astral.String8
	LastSeen    astral.Time
	Attachments *astral.Bundle
}

// astral

func (Broadcaster) ObjectType() string {
	return "mod.nearby.broadcaster"
}

func (b Broadcaster) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(b).WriteTo(w)
}

func (b *Broadcaster) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(b).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&Broadcaster{})
}
