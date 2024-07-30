package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"io"
	"strings"
	"time"
)

const RouteServiceName = ".gateway"
const RouteServiceType = "mod.gateway.route"

const acceptTimeout = 15 * time.Second

type RouteService struct {
	*Module
	router astral.Router
}

func (srv *RouteService) Run(ctx context.Context) error {
	err := srv.AddRoute(RouteServiceName+".*", srv)
	if err != nil {
		return err
	}
	defer srv.RemoveRoute(RouteServiceName + ".*")

	<-ctx.Done()
	return nil
}

func (srv *RouteService) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	var targetKey string

	switch {
	case strings.HasPrefix(query.Query, RouteServiceName+"."):
		targetKey, _ = strings.CutPrefix(query.Query, RouteServiceName+".")

	default:
		return astral.Reject()
	}

	// check if the target is us
	if targetKey == srv.node.Identity().PublicKeyHex() {
		return astral.Accept(query, caller, func(conn astral.Conn) {
			gwConn := newConn(
				conn,
				NewEndpoint(query.Target, query.Target),
				NewEndpoint(query.Caller, query.Target),
				false,
			)

			actx, cancel := context.WithTimeout(context.Background(), acceptTimeout)
			defer cancel()

			err := srv.Nodes.Accept(actx, gwConn)
			if err != nil {
				return
			}
		})
	}

	targetIdentity, err := id.ParsePublicKeyHex(targetKey)
	if err != nil {
		return astral.Reject()
	}

	nextQuery := &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: srv.node.Identity(),
		Target: targetIdentity,
		Query:  query.Query,
	}

	srv.log.Logv(2, "forwarding %v to %v", query.Caller, targetIdentity)

	return srv.router.RouteQuery(ctx, nextQuery, caller)
}
