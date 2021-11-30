package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

type Resolver interface {
	Resolve(nodeID id.Identity) <-chan infra.Addr
}
