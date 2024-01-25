package relay

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/relay/proto"
	"github.com/cryptopunkscc/astrald/net"
)

// RouteVia routes a query through a relay
func (mod *Module) RouteVia(
	ctx context.Context,
	relayID id.Identity,
	query net.Query,
	caller net.SecureWriteCloser,
	hints net.Hints,
) (target net.SecureWriteCloser, err error) {
	if hints.Origin != net.OriginLocal {
		return net.RouteNotFound(mod, errors.New("not a local query"))
	}

	// prepare query parameters
	var queryParams = &proto.QueryParams{
		Target: query.Target(),
		Query:  query.Query(),
		Nonce:  uint64(query.Nonce()),
	}

	if relayID.IsEqual(mod.node.Identity()) {
		return net.RouteNotFound(mod, errors.New("cannot relay via localnode"))
	}

	// check if relaying makes sense (local and/or remote identity needs to be changed)
	var callerEqual = query.Caller().IsEqual(mod.node.Identity())
	var targetEqual = query.Target().IsEqual(relayID)
	if callerEqual && targetEqual {
		return net.RouteNotFound(mod, errors.New("relay not needed"))
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
			return net.RouteNotFound(mod, err)
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
	relayConn, err := net.RouteWithHints(
		ctx,
		mod.node.Router(),
		net.NewQuery(mod.node.Identity(), relayID, relay.RelayServiceName),
		net.DefaultHints().SetSilent(),
	)
	if err != nil {
		return net.RouteNotFound(mod, fmt.Errorf("remote relay service unavailable (%w)", err))
	}
	defer relayConn.Close()

	var relayService = proto.New(relayConn)

	// query the relay
	response, err := relayService.Query(queryParams)
	switch {
	case errors.Is(err, proto.ErrRejected):
		return net.Reject()
	case errors.Is(err, proto.ErrRouteNotFound):
		return net.RouteNotFound(mod)
	case err != nil:
		return net.RouteNotFound(mod, err)
	}

	var targetIM = NewIdentityMachine(relayID)

	// apply target certificate
	if len(response.Cert) > 0 {
		if err = targetIM.Apply(response.Cert); err != nil {
			return net.RouteNotFound(mod, err)
		}
	}

	// verify target identity
	if !targetIM.Identity().IsEqual(query.Target()) {
		return net.RouteNotFound(mod, errors.New("target identity mismatch"))
	}

	// route through the proxy service
	var proxyQuery = net.NewQueryNonce(mod.node.Identity(), relayID, response.ProxyService, query.Nonce())
	if !caller.Identity().IsEqual(mod.node.Identity()) {
		caller = net.NewIdentityTranslation(caller, mod.node.Identity())
	}
	proxy, err := mod.node.Router().RouteQuery(ctx, proxyQuery, caller, net.DefaultHints().SetReroute())
	if err != nil {
		return nil, err
	}

	if !proxy.Identity().IsEqual(query.Target()) {
		proxy = net.NewIdentityTranslation(proxy, query.Target())
	}

	return proxy, nil
}

func (mod *Module) RouterFuncVia(relay id.Identity) net.RouteQueryFunc {
	return func(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
		return mod.RouteVia(ctx, relay, query, caller, hints)
	}
}
