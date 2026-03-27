package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	ModuleName = "nearby"
)

type Module interface {
	Broadcast() error
	AddStatusComposer(Composer)
	Mode() Mode
	SetMode(ctx *astral.Context, m Mode) error
	ResolveStatus(status *StatusMessage) *astral.Identity
}

type Composer interface {
	ComposeStatus(Composition)
}

type Composition interface {
	Receiver() *astral.Identity
	Attach(astral.Object) error
}
