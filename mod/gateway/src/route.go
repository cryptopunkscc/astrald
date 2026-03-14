package gateway

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

const acceptTimeout = 30 * time.Second

func (mod *Module) routeQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	var targetKey string
	switch {
	case strings.HasPrefix(q.Query, gateway.MethodRoute+"."):
		targetKey, _ = strings.CutPrefix(q.Query, gateway.MethodRoute+".")
	default:
		return query.Reject()
	}

	// target is us
	if targetKey == mod.node.Identity().String() {
		return query.Accept(q, w, func(conn astral.Conn) {
			c := &gwConn{
				ReadWriteCloser: conn,
				local:           gateway.NewEndpoint(q.Target, q.Target),
				remote:          gateway.NewEndpoint(q.Caller, q.Target),
				outbound:        false,
			}

			// prevents slow gateway connections
			actx, cancel := context.WithTimeout(context.Background(), acceptTimeout)
			defer cancel()

			if err := mod.Nodes.EstablishInboundLink(actx, c); err != nil {
				mod.log.Errorv(1, "inbound link from %v failed: %v", q.Caller, err)
			}
		})
	}

	// forward query (will automatically use existing link)

	targetIdentity, err := astral.ParseIdentity(targetKey)
	if err != nil {
		return query.Reject()
	}

	nextQuery := &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: mod.node.Identity(),
		Target: targetIdentity,
		Query:  q.Query,
	}

	mod.log.Logv(2, "routing %v to %v via link", q.Caller, targetIdentity)
	return mod.node.RouteQuery(ctx, nextQuery, w)
}
