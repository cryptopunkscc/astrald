package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
)

const serviceName = "connect"

type Connect struct {
	node node.Node
}

func (mod *Connect) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		l, err := link.Accept(ctx, conn, mod.node.Identity())
		if err != nil {
			return
		}

		mod.node.Network().AddLink(l)
	})
}

func (mod *Connect) Run(ctx context.Context) error {
	_, err := mod.node.Services().Register(ctx, mod.node.Identity(), serviceName, mod)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}
