package dir

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
)

const ModuleName = "dir"

type Module interface {
	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error
}

type Describer desc.Describer[id.Identity]
