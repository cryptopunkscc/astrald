package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceHandle = "connect"

type Connect struct {
	node *node.Node
}

func (mod *Connect) Run(ctx context.Context) error {
	port, err := mod.node.Ports.RegisterContext(ctx, serviceHandle)
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

		mod.node.Peers.Server.Conns <- &wrapper{
			local:           mod.node.Identity().Public(),
			remote:          query.Link().RemoteIdentity(),
			ReadWriteCloser: conn,
			outbound:        false,
		}
	}

	return nil
}
