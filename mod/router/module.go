package router

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/router/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"time"
)

type Module struct {
	node   node.Node
	keys   assets.KeyStore
	log    *log.Logger
	config Config
	ctx    context.Context
}

func (m *Module) Run(ctx context.Context) error {
	m.ctx = ctx

	// register as a router
	if coreRouter, ok := m.node.Router().(*node.CoreRouter); ok {
		coreRouter.Routers.AddRouter(m)
	} else {
		return errors.New("unsupported router type")
	}

	return tasks.Group(
		&RouterService{Module: m},
	).Run(ctx)
}

func (m *Module) RouteVia(
	ctx context.Context,
	routerIdentity id.Identity,
	query net.Query,
	caller net.SecureWriteCloser,
	hints net.Hints,
) (target net.SecureWriteCloser, err error) {
	// TODO: remove this once we have persistent certificates
	if query.Caller().PrivateKey() == nil {
		return net.RouteNotFound(m, errors.New("caller private key missing"))
	}

	// prepare query parameters
	var queryParams = &proto.QueryParams{
		Target: query.Target(),
		Query:  query.Query(),
	}

	// attach a caller certificate if necessary
	if !query.Caller().IsEqual(m.node.Identity()) {
		// TODO: fetch certificate from db instead of signing a new one every time
		var cert = NewRouterCert(query.Caller(), m.node.Identity(), time.Now().Add(time.Minute))
		queryParams.Cert, err = cslq.Marshal(cert)
		if err != nil {
			return net.RouteNotFound(m, err)
		}
	}

	// open a router session
	routerConn, err := net.Route(ctx, m.node.Router(), net.NewQuery(m.node.Identity(), routerIdentity, RouterServiceName))
	if err != nil {
		return nil, err
	}
	defer routerConn.Close()
	var router = proto.New(routerConn)

	// query the router
	response, err := router.Query(queryParams)
	switch {
	case errors.Is(err, proto.ErrRejected):
		return net.Reject()
	case err != nil:
		return net.RouteNotFound(m, err)
	}

	var targetIM = NewIdentityMachine(routerIdentity)

	// apply target certificate
	if len(response.Cert) > 0 {
		if err = targetIM.Apply(response.Cert); err != nil {
			return net.RouteNotFound(m, err)
		}
	}

	// verify target identity
	if !targetIM.Identity().IsEqual(query.Target()) {
		return net.RouteNotFound(m, errors.New("target identity mismatch"))
	}

	// route through the proxy service
	var proxyQuery = net.NewQuery(m.node.Identity(), routerIdentity, response.ProxyService)
	if !caller.Identity().IsEqual(m.node.Identity()) {
		caller = NewIdentityTranslation(caller, m.node.Identity())
	}
	proxy, err := m.node.Router().RouteQuery(ctx, proxyQuery, caller, net.DefaultHints().SetDontMonitor().SetAllowRedirect())
	if err != nil {
		return net.RouteNotFound(m, err)
	}

	if !proxy.Identity().IsEqual(query.Target()) {
		proxy = NewIdentityTranslation(proxy, query.Target())
	}

	return proxy, nil
}

func (m *Module) isLocal(identity id.Identity) bool {
	if m.node.Identity().IsEqual(identity) {
		return true
	}
	for _, info := range m.node.Services().List() {
		if info.Identity.IsEqual(identity) {
			return true
		}
	}
	return false
}

func (m *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if m.isLocal(query.Target()) {
		return net.RouteNotFound(m)
	}

	if query.Caller().IsEqual(m.node.Identity()) {
		return net.RouteNotFound(m)
	}

	if m.node.Network().Links().ByRemoteIdentity(query.Target()).Count() > 0 {
		return m.RouteVia(ctx, query.Target(), query, caller, hints)
	}

	return net.RouteNotFound(m)
}
