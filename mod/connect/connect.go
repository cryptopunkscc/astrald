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

func (mod *Connect) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		newConn, err := mod.node.Network().Server().Handshake(ctx, conn)
		if err != nil {
			return
		}

		mod.node.Network().AddLink(link.NewCoreLink(newConn))
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
