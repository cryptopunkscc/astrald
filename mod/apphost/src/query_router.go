package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	guest, ok := mod.guests.Get(q.Target.String())
	if !ok {
		return query.RouteNotFound(mod)
	}

	return NewRelay(guest.Endpoint, guest.Token, mod.log).RouteQuery(ctx, q, w)
}
