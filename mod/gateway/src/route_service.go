package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"time"
)

const RouteServiceName = ".gateway"
const RouteServiceType = "mod.gateway.route"

const acceptTimeout = 15 * time.Second

type RouteService struct {
	*Module
	router net.Router
}

func (srv *RouteService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(RouteServiceName+".*", srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(RouteServiceName + ".*")

	if srv.sdp != nil {
		srv.sdp.AddServiceDiscoverer(srv)
		defer srv.sdp.RemoveServiceDiscoverer(srv)
	}

	<-ctx.Done()
	return nil
}

func (srv *RouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var targetKey string

	switch {
	case strings.HasPrefix(query.Query(), RouteServiceName+"."):
		targetKey, _ = strings.CutPrefix(query.Query(), RouteServiceName+".")

	default:
		return net.Reject()
	}

	// check if the target is us
	if targetKey == srv.node.Identity().PublicKeyHex() {
		return net.Accept(query, caller, func(conn net.SecureConn) {
			gwConn := newConn(
				conn,
				NewEndpoint(query.Target(), query.Target()),
				NewEndpoint(query.Caller(), query.Target()),
				false,
			)

			actx, cancel := context.WithTimeout(context.Background(), acceptTimeout)
			defer cancel()

			_, err := srv.nodes.AcceptLink(actx, gwConn)
			if err != nil {
				return
			}
		})
	}

	targetIdentity, err := id.ParsePublicKeyHex(targetKey)
	if err != nil {
		return net.Reject()
	}

	maskedQuery := net.NewQueryNonce(
		srv.node.Identity(),
		targetIdentity,
		query.Query(),
		query.Nonce(),
	)

	maskedCaller := net.NewIdentityTranslation(caller, srv.node.Identity())

	srv.log.Logv(2, "forwarding %v to %v", query.Caller(), targetIdentity)

	dst, err := srv.router.RouteQuery(ctx, maskedQuery, maskedCaller, hints.SetReroute())
	if err != nil {
		return nil, err
	}

	var maskedTarget = net.NewIdentityTranslation(dst, srv.node.Identity())

	return maskedTarget, nil
}

func (srv *RouteService) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	return []discovery.Service{
		{
			Identity: srv.node.Identity(),
			Name:     RouteServiceName,
			Type:     RouteServiceType,
		},
	}, nil
}
