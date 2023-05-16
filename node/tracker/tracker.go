package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type Tracker interface {
	Add(identity id.Identity, endpoint net.Endpoint, expiresAt time.Time) error
	FindAll(identity id.Identity) ([]TrackedEndpoint, error)
	DeleteAll(identity id.Identity) error
	Identities() ([]id.Identity, error)
}
