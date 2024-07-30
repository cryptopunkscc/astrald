package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if !query.Target.IsEqual(mod.UserID()) {
		return astral.RouteNotFound(mod)
	}

	return mod.PathRouter.RouteQuery(ctx, query, caller)
}
