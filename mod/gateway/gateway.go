package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/streams"
)

const ModuleName = "gateway"
const queryConnect = "connect"

type Gateway struct{}

func (Gateway) Run(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.RegisterContext(ctx, gw.PortName)
	if err != nil {
		return err
	}

	for req := range port.Queries() {
		conn := req.Accept()

		go func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var cookie string

			err = cslq.Decode(conn, "[c]c", &cookie)
			if err != nil {
				conn.Close()
				return
			}

			nodeID, err := id.ParsePublicKeyHex(cookie)
			if err != nil {
				conn.Close()
				return
			}

			out, err := node.Query(ctx, nodeID, queryConnect)
			if err != nil {
				cslq.Encode(conn, "c", false)
				conn.Close()
				return
			}

			cslq.Encode(conn, "c", true)

			go func() {
				<-ctx.Done()
				out.Close()
			}()

			streams.Join(conn, out)
		}()
	}

	return nil
}

func (Gateway) String() string {
	return ModuleName
}
