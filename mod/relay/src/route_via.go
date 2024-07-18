package relay

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/relay/proto"
	"github.com/cryptopunkscc/astrald/astral"
)

// RouteVia routes a query through a relay
func (mod *Module) RouteVia(
	ctx context.Context,
	relayID id.Identity,
	query astral.Query,
	caller astral.SecureWriteCloser,
	hints astral.Hints,
) (target astral.SecureWriteCloser, err error) {
	if hints.Origin != astral.OriginLocal {
		return astral.RouteNotFound(mod, errors.New("not a local query"))
	}

	// prepare query parameters
	var queryParams = &proto.QueryParams{
		Target: query.Target(),
		Query:  query.Query(),
		Nonce:  uint64(query.Nonce()),
	}

	if relayID.IsEqual(mod.node.Identity()) {
		return astral.RouteNotFound(mod, errors.New("cannot relay via localnode"))
	}

	// check if relaying makes sense (local and/or remote identity needs to be changed)
	var callerEqual = query.Caller().IsEqual(mod.node.Identity())
	var targetEqual = query.Target().IsEqual(relayID)
	if callerEqual && targetEqual {
		return astral.RouteNotFound(mod, errors.New("relay not needed"))
	}

	// attach a caller certificate if necessary
	if !query.Caller().IsEqual(mod.node.Identity()) {
		queryParams.Cert, err = mod.ReadCert(
			&relay.FindOpts{
				TargetID:  query.Caller(),
				RelayID:   mod.node.Identity(),
				Direction: relay.Outbound,
			},
		)
		if err != nil {
			mod.log.Errorv(1, "error getting caller certificate: %v", err)
			return astral.RouteNotFound(mod, err)
		}
	}

	mod.log.Logv(2, "[%v] %v@%v -> %v@%v:%v",
		query.Nonce(),
		query.Caller(),
		mod.node.Identity(),
		query.Target(),
		relayID,
		query.Query(),
	)

	// open a relay session
	relayConn, err := astral.RouteWithHints(
		ctx,
		mod.node.Router(),
		astral.NewQuery(mod.node.Identity(), relayID, relay.ServiceName),
		astral.DefaultHints().SetSilent(),
	)
	if err != nil {
		return astral.RouteNotFound(mod, fmt.Errorf("remote relay service unavailable (%w)", err))
	}
	defer relayConn.Close()

	var relayService = proto.New(relayConn)

	// query the relay
	response, err := relayService.Query(queryParams)
	switch {
	case errors.Is(err, proto.ErrRejected):
		return astral.Reject()
	case errors.Is(err, proto.ErrRouteNotFound):
		return astral.RouteNotFound(mod)
	case err != nil:
		return astral.RouteNotFound(mod, err)
	}

	var targetIM = NewIdentityMachine(relayID)

	// apply target certificate
	if len(response.Cert) > 0 {
		if err = targetIM.Apply(response.Cert); err != nil {
			return astral.RouteNotFound(mod, err)
		}
	}

	// verify target identity
	if !targetIM.Identity().IsEqual(query.Target()) {
		return astral.RouteNotFound(mod, errors.New("target identity mismatch"))
	}

	// route through the proxy service
	var proxyQuery = astral.NewQueryNonce(mod.node.Identity(), relayID, response.ProxyService, query.Nonce())
	if !caller.Identity().IsEqual(mod.node.Identity()) {
		caller = astral.NewIdentityTranslation(caller, mod.node.Identity())
	}
	proxy, err := mod.node.Router().RouteQuery(ctx, proxyQuery, caller, astral.DefaultHints().SetReroute())
	if err != nil {
		return nil, err
	}

	if !proxy.Identity().IsEqual(query.Target()) {
		proxy = astral.NewIdentityTranslation(proxy, query.Target())
	}

	return proxy, nil
}

func (mod *Module) RouterFuncVia(relay id.Identity) astral.RouteQueryFunc {
	return func(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
		return mod.RouteVia(ctx, relay, query, caller, hints)
	}
}
