package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"io"
	"strings"
	"time"
)

const RouteServiceName = ".gateway"
const acceptTimeout = 15 * time.Second

type RouteService struct {
	*Module
	router astral.Router
}

func (srv *RouteService) Run(ctx *astral.Context) error {
	err := srv.AddRoute(RouteServiceName+".*", srv)
	if err != nil {
		return err
	}
	defer srv.RemoveRoute(RouteServiceName + ".*")

	<-ctx.Done()
	return nil
}

func (srv *RouteService) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var targetKey string

	switch {
	case strings.HasPrefix(q.Query, RouteServiceName+"."):
		targetKey, _ = strings.CutPrefix(q.Query, RouteServiceName+".")

	default:
		return query.Reject()
	}

	// check if the target is us
	if targetKey == srv.node.Identity().String() {
		return query.Accept(q, w, func(conn astral.Conn) {
			gwConn := newConn(
				conn,
				gateway.NewEndpoint(q.Target, q.Target),
				gateway.NewEndpoint(q.Caller, q.Target),
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

	targetIdentity, err := astral.IdentityFromString(targetKey)
	if err != nil {
		return query.Reject()
	}

	nextQuery := &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: srv.node.Identity(),
		Target: targetIdentity,
		Query:  q.Query,
	}

	srv.log.Logv(2, "forwarding %v to %v", q.Caller, targetIdentity)

	return srv.router.RouteQuery(ctx, nextQuery, w)
}
