package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !query.Target().IsEqual(mod.UserID()) {
		return net.RouteNotFound(mod)
	}

	return mod.PathRouter.RouteQuery(ctx, query, caller, hints)
}
