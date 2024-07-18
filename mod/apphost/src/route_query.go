package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	if guest := mod.getGuest(query.Target()); guest != nil {
		return guest.RouteQuery(ctx, query, caller, hints)
	}

	return astral.RouteNotFound(mod)
}
