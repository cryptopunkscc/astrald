package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	query "github.com/cryptopunkscc/astrald/query"
)

const serviceName = "connect"

type Connect struct {
	node node.Node
}

func (mod *Connect) RouteQuery(ctx context.Context, q query.Query, swc net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return query.Accept(q, swc, func(conn net.SecureConn) {
		newConn, err := mod.node.Network().Server().Handshake(ctx, conn)
		if err != nil {
			return
		}

		mod.node.Network().AddSecureConn(newConn)
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
