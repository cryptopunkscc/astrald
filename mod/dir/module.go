package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "dir"
const DBPrefix = "dir__"

type Module interface {
	SetAlias(*astral.Identity, string) error
	GetAlias(*astral.Identity) (string, error)
	ResolveIdentity(string) (*astral.Identity, error)
	DisplayName(*astral.Identity) string
	AddResolver(Resolver) error
}

type Resolver interface {
	ResolveIdentity(string) (*astral.Identity, error)
	DisplayName(*astral.Identity) string
}
