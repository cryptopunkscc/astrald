package relay

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	if mod.isLocal(query.Target()) {
		return astral.RouteNotFound(mod, errors.New("target is local node"))
	}

	relays, err := mod.FindExternalRelays(query.Target())
	if err != nil {
		return astral.RouteNotFound(mod, err)
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

	return astral.RouteNotFound(mod, errs...)
}
