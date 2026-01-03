package core

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
)

// routerAdapter is an adapter that allows lib/astrald client libraries to use an astral.Router directly
type routerAdapter struct {
	astral.Router
	identity *astral.Identity
}

var _ astrald.Router = &routerAdapter{}

func (r *routerAdapter) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	return query.Route(ctx, r.Router, q)
}

func (r *routerAdapter) GuestID() *astral.Identity {
	return r.identity
}

func (r *routerAdapter) HostID() *astral.Identity {
	return r.identity
}
