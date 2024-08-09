package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

var _ astral.Router = &Module{}

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if mod.UserID().IsZero() {
		return query.RouteNotFound(mod)
	}

	if !q.Target.IsEqual(mod.UserID()) {
		return query.RouteNotFound(mod)
	}

	return mod.PathRouter.RouteQuery(ctx, q, w)
}
