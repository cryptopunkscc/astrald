package dir

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
)

const ModuleName = "dir"
const DBPrefix = "dir__"

type Module interface {
	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	SetAlias(id.Identity, string) error
	GetAlias(id.Identity) (string, error)
	Resolve(string) (id.Identity, error)
}

type Describer desc.Describer[id.Identity]
