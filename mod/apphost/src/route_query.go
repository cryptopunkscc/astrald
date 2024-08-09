package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if r, err := mod.router.RouteQuery(ctx, q, w); err == nil {
		return r, nil
	}

	if guest := mod.getGuest(q.Target); guest != nil {
		return guest.RouteQuery(ctx, q, w)
	}

	return query.RouteNotFound(mod)
}
