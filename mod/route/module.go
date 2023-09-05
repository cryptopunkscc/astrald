package route

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/route/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
)

type Module struct {
	node   node.Node
	keys   assets.KeyStore
	log    *log.Logger
	config Config
}

func (m *Module) Run(ctx context.Context) error {
	// register as a router
	if coreRouter, ok := m.node.Router().(*node.CoreRouter); ok {
		coreRouter.Routers.AddRouter(m)
	}

	return tasks.Group(
		&RouteService{Module: m},
	).Run(ctx)
}

func (m *Module) RouteVia(ctx context.Context, relay id.Identity, query net.Query, caller net.SecureWriteCloser) (target net.SecureWriteCloser, err error) {
	if query.Caller().PrivateKey() == nil {
		return nil, errors.New("caller private key missing")
	}

	// call the router on the relay
	routeConn, err := net.Route(ctx, m.node.Router(), net.NewQuery(m.node.Identity(), relay, RouteServiceName))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			routeConn.Close()
		}
	}()

	var rpc = proto.New(routeConn)

	// present a certificate if needed
	if !query.Caller().IsEqual(m.node.Identity()) {
		var cert = proto.NewRelayCert(query.Caller(), routeConn.LocalIdentity())

		if err = rpc.Shift(cert); err != nil {
			return nil, err
		}
	}

	// send the query
	err = rpc.Query(query.Target(), query.Query())
	if err != nil {
		if errors.Is(err, proto.ErrRejected) {
			return net.Reject()
		}
		return nil, &net.ErrRouteNotFound{Router: m}
	}

	// expect a certificate if the remote party needs to provide it
	if !routeConn.RemoteIdentity().IsEqual(query.Target()) {
		var cert proto.RelayCert
		if err := rpc.Decode(&cert); err != nil {
			return nil, err
		}

		if !cert.Identity.IsEqual(query.Target()) {
			return nil, errors.New("received invalid certificate")
		}

		if !cert.Relay.IsEqual(routeConn.RemoteIdentity()) {
			return nil, errors.New("received invalid certificate")
		}

		if err = cert.Verify(); err != nil {
			return nil, err
		}

		m.log.Logv(2, "target shifted to %v", cert.Identity)
	}

	if err = rpc.DecodeErr(); err != nil {
		return nil, err
	}

	go func() {
		io.Copy(caller, routeConn)
		caller.Close()
	}()

	return net.NewSecureWriteCloser(routeConn, query.Target()), nil
}

func (m *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if query.Caller().IsEqual(m.node.Identity()) {
		return nil, &net.ErrRouteNotFound{Router: m}
	}

	if m.node.Network().Links().ByRemoteIdentity(query.Target()).Count() > 0 {
		return m.RouteVia(ctx, query.Target(), query, caller)
	}

	return nil, &net.ErrRouteNotFound{Router: m}
}
