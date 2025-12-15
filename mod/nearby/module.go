package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	ModuleName = "nearby"
)

type Module interface {
	Broadcast() error
	Broadcasters() []*Broadcaster
	AddStatusComposer(Composer)
}

type Composer interface {
	ComposeStatus(Composition)
}

type Composition interface {
	Receiver() *astral.Identity
	Attach(astral.Object) error
}
