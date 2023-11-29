package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"strings"
)

const RouteServiceName = "gateway.route"
const RouteServiceType = "gateway.route"

type RouteService struct {
	*Module
	router net.Router
}

func (srv *RouteService) Run(ctx context.Context) error {
	var err = srv.node.AddRoute(RouteServiceName+"?*", srv)
	if err != nil {
		return err
	}
	defer srv.node.RemoveRoute(RouteServiceName + "?*")

	err = srv.node.AddRoute(RouteServiceName+".*", srv)
	if err != nil {
		return err
	}
	defer srv.node.RemoveRoute(RouteServiceName + ".*")

	if srv.sdp != nil {
		srv.sdp.AddSource(srv)
		defer srv.sdp.RemoveSource(srv)
	}

	<-ctx.Done()
	return nil
}

func (srv *RouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var targetKey string

	switch {
	case strings.HasPrefix(query.Query(), RouteServiceName+"."):
		targetKey, _ = strings.CutPrefix(query.Query(), RouteServiceName+".")

	case strings.HasPrefix(query.Query(), RouteServiceName+"?"):
		targetKey, _ = strings.CutPrefix(query.Query(), RouteServiceName+"?")

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
			l, err := link.Accept(ctx, gwConn, srv.node.Identity())
			if err != nil {
				return
			}

			err = srv.node.Network().AddLink(l)
			if err != nil {
				l.Close()
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

func (srv *RouteService) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	return []sdp.ServiceEntry{
		{
			Identity: srv.node.Identity(),
			Name:     RouteServiceName,
			Type:     RouteServiceType,
		},
	}, nil
}
