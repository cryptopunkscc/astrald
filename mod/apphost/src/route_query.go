package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if guest := mod.getGuest(query.Target); guest != nil {
		return guest.RouteQuery(ctx, query, caller)
	}

	return astral.RouteNotFound(mod)
}
