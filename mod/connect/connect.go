package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceHandle = "connect"

type Connect struct {
	node node.Node
}

func (mod *Connect) Run(ctx context.Context) error {
	port, err := mod.node.Services().RegisterContext(ctx, serviceHandle, mod.node.Identity())
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		// skip local queries
		if query.IsLocal() {
			query.Reject()
			continue
		}

		conn, err := query.Accept()
		if err != nil {
			continue
		}

		infraConn := &wrapper{
			local:           mod.node.Identity().Public(),
			remote:          query.Link().RemoteIdentity(),
			ReadWriteCloser: conn,
			outbound:        false,
		}

		authConn, err := mod.node.Network().Server().Handshake(ctx, infraConn)
		if err != nil {
			infraConn.Close()
			continue
		}

		mod.node.Network().AddSecureConn(authConn)
	}

	return nil
}
