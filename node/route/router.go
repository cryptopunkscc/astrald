package route

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

type Router interface {
	Route(nodeID id.Identity) *Route
	AddRoute(r *Route)
	AddAddr(nodeID id.Identity, addr infra.Addr)
}
