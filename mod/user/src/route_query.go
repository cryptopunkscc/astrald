package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	if !query.Target().IsEqual(mod.UserID()) {
		return astral.RouteNotFound(mod)
	}

	return mod.PathRouter.RouteQuery(ctx, query, caller, hints)
}
