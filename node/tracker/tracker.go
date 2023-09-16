package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type Tracker interface {
	AddEndpoint(identity id.Identity, endpoint net.Endpoint) error
	EndpointsByIdentity(identity id.Identity) ([]net.Endpoint, error)
	Clear(identity id.Identity) error
	Remove(identity id.Identity) error
	Identities() ([]id.Identity, error)
	SetAlias(identity id.Identity, alias string) error
	GetAlias(identity id.Identity) (string, error)
	IdentityByAlias(alias string) (id.Identity, error)
}
