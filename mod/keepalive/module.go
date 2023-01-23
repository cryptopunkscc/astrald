package keepalive

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const portName = "net.keepalive"

type Module struct {
	*node.Node
}

var log = _log.Tag(ModuleName)

func (m *Module) Run(ctx context.Context) error {
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
		log.Log("timeout disabled for %s over %s",
			conn.RemoteIdentity(),
			conn.Link().Network(),
		)
	}

	return nil
}
