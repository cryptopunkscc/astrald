package route

import "github.com/cryptopunkscc/astrald/auth/id"

type Router interface {
	Route(nodeID id.Identity) *Route
	AddRoute(r *Route)
}
