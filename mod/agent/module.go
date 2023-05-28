package agent

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const portName = "sys.agent"

type Module struct {
	node node.Node
}

var log = _log.Tag(ModuleName)

func (m *Module) Run(ctx context.Context) error {
	port, err := m.node.Services().RegisterContext(ctx, portName, m.node.Identity())
	if err != nil {
		return err
	}

	for q := range port.Queries() {
		// allow local connections only
		if !q.IsLocal() {
			q.Reject()
			continue
		}

		conn, err := q.Accept()
		if err != nil {
			continue
		}

		s := &Server{
			node: m.node,
			conn: conn,
		}

		go func() {
			if err := s.Run(ctx); err != nil {
				log.Error("serve error: %s", err)
			}
		}()
	}

	return nil
}
