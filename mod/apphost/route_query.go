package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if guest := mod.getGuest(query.Target()); guest != nil {
		return guest.RouteQuery(ctx, query, caller, hints)
	}

	return net.RouteNotFound(mod)
}
