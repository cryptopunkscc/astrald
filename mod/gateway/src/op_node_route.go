package gateway

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type opNodeRouteArgs struct {
	Target *astral.Identity
}

func (mod *Module) OpNodeRoute(ctx *astral.Context, q *ops.Query, args opNodeRouteArgs) error {
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	// target is this node — accept and establish inbound link
	if args.Target.IsEqual(mod.node.Identity()) {
		conn := q.Accept()
		c := &gatewayConn{
			ReadWriteCloser: conn,
			local:           gateway.NewEndpoint(q.Target, q.Target),
			remote:          gateway.NewEndpoint(q.Caller(), q.Target),
		}

		actx, cancel := context.WithTimeout(context.Background(), acceptTimeout)
		defer cancel()

		if err := mod.Nodes.EstablishInboundLink(actx, c); err != nil {
			mod.log.Errorv(1, "inbound link from %v failed: %v", q.Caller(), err)
		}
		return nil
	}

	// forward: accept caller side, dial target side, pipe
	inConn := q.Accept()
	nextQ := query.New(mod.node.Identity(), args.Target, gateway.MethodNodeRoute, query.Args{"target": args.Target})
	outConn, err := query.Route(ctx, mod.node, astral.Launch(nextQ))
	if err != nil {
		inConn.Close()
		return err
	}

	go pipe(inConn, outConn)
	return nil
}
