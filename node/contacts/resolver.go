package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Resolver interface {
	Lookup(nodeID id.Identity) (<-chan *Addr, error)
}
