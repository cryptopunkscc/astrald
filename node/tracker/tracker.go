package tracker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type Tracker interface {
	Add(identity id.Identity, addr net.Endpoint, expiresAt time.Time) error
	Identities() ([]id.Identity, error)
	ForgetIdentity(identity id.Identity) error
	Watch(ctx context.Context, nodeID id.Identity) <-chan *Addr
	AddrByIdentity(identity id.Identity) ([]Addr, error)
}
