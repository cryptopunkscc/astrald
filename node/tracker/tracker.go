package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type Tracker interface {
	AddEndpoint(identity id.Identity, endpoint net.Endpoint, expiresAt time.Time) error
	EndpointsByIdentity(identity id.Identity) ([]TrackedEndpoint, error)
	DeleteAll(identity id.Identity) error
	Identities() ([]id.Identity, error)
	SetAlias(identity id.Identity, alias string) error
	GetAlias(identity id.Identity) (string, error)
	IdentityByAlias(alias string) (id.Identity, error)
}
