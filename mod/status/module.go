package status

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	ModuleName = "status"
)

type Module interface {
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
