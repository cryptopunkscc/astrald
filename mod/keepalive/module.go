package keepalive

import (
	"context"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"log"
)

const portName = "net.keepalive"

type Module struct {
	*node.Node
}

func (m Module) Run(ctx context.Context) error {
	port, err := m.Ports.RegisterContext(ctx, portName)
	if err != nil {
		return err
	}

	for q := range port.Queries() {
		if q.IsLocal() {
			q.Reject()
			continue
		}

		conn, err := q.Accept()
		if err == nil {
			conn.Close()
		}

		// disable timeout on the link
		conn.Link().SetIdleTimeout(0)
		log.Printf("[keepalive] timeout disabled for %s link with %s\n",
			conn.Link().Network(),
			logfmt.ID(conn.RemoteIdentity()),
		)
	}

	return nil
}
