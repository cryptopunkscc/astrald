package connect

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceHandle = ".connect"

func connect(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.RegisterContext(ctx, serviceHandle)
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		conn := query.Accept()

		node.Server.Conns <- &wrapper{
			local:           node.Identity().Public(),
			remote:          query.Caller(),
			ReadWriteCloser: conn,
			outbound:        false,
		}
	}

	return nil
}

func init() {
	_ = node.RegisterService(serviceHandle, connect)
}
