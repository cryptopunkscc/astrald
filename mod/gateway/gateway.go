package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/streams"
)

const queryConnect = "connect"

type Gateway struct {
	node *node.Node
}

func (mod *Gateway) Run(ctx context.Context) error {
	port, err := mod.node.Ports.RegisterContext(ctx, gw.PortName)
	if err != nil {
		return err
	}

	for req := range port.Queries() {
		conn, err := req.Accept()
		if err != nil {
			continue
		}

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

			out, err := mod.node.Query(ctx, nodeID, queryConnect)
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
