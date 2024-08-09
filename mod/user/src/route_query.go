package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !q.Target.IsEqual(mod.UserID()) {
		return astral.RouteNotFound(mod)
	}

	return mod.PathRouter.RouteQuery(ctx, q, w)
}
