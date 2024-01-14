package relay

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if mod.isLocal(query.Target()) {
		return net.RouteNotFound(mod, errors.New("target is local node"))
	}

	relays, err := mod.FindExternalRelays(query.Target())
	if err != nil {
		return net.RouteNotFound(mod, err)
	}

	var errs []error
	for _, relay := range relays {
		target, err := mod.RouteVia(ctx, relay, query, caller, hints)
		if err == nil {
			return target, nil
		}
		errs = append(errs, err)
	}

	if !query.Caller().IsEqual(mod.node.Identity()) {
		return mod.RouteVia(ctx, query.Target(), query, caller, hints)
	}

	return net.RouteNotFound(mod, errs...)
}
