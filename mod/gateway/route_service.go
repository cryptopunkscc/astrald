package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/router"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/modules"
	"strings"
)

const RouteServiceName = "gateway.route"
const RouteServiceType = "gateway.route"

type RouteService struct {
	*Module
	router net.Router
}

func (srv *RouteService) Run(ctx context.Context) error {
	var err = srv.node.AddRoute(RouteServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.RemoveRoute(RouteServiceName)

	if disco, err := modules.Find[*sdp.Module](srv.node.Modules()); err == nil {
		disco.AddSource(srv)
		defer disco.RemoveSource(srv)
	}

	<-ctx.Done()
	return nil
}

func (srv *RouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, arg, ok := strings.Cut(query.Query(), "?")
	if !ok {
		return net.Reject()
	}

	// check if the target is us
	if arg == srv.node.Identity().PublicKeyHex() {
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

	targetIdentity, err := id.ParsePublicKeyHex(arg)
	if err != nil {
		return net.Reject()
	}

	maskedQuery := net.NewQueryNonce(
		srv.node.Identity(),
		targetIdentity,
		query.Query(),
		query.Nonce(),
	)

	maskedCaller := router.NewIdentityTranslation(caller, srv.node.Identity())

	dst, err := srv.router.RouteQuery(ctx, maskedQuery, maskedCaller, hints.SetReroute())
	if err != nil {
		return nil, err
	}

	var maskedTarget = router.NewIdentityTranslation(dst, srv.node.Identity())

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
