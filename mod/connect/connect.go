package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceHandle = "connect"
const ModuleName = "connect"

type Connect struct{}

func (Connect) Run(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.RegisterContext(ctx, serviceHandle)
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

		node.Peers.Server.Conns <- &wrapper{
			local:           node.Identity().Public(),
			remote:          query.Link().RemoteIdentity(),
			ReadWriteCloser: conn,
			outbound:        false,
		}
	}

	return nil
}

func (Connect) String() string {
	return ModuleName
}
