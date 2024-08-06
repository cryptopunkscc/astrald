package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/desc"
)

const ModuleName = "dir"
const DBPrefix = "dir__"

type Module interface {
	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	SetAlias(*astral.Identity, string) error
	GetAlias(*astral.Identity) (string, error)
	Resolve(string) (*astral.Identity, error)
	DisplayName(identity *astral.Identity) string
	AddResolver(Resolver) error
}

type Describer desc.Describer[*astral.Identity]

type Resolver interface {
	Resolve(s string) (*astral.Identity, error)
	DisplayName(identity *astral.Identity) string
}
