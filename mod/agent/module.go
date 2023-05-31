package agent

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
)

const portName = "sys.agent"

type Module struct {
	node node.Node
}

var log = _log.Tag(ModuleName)

func (m *Module) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(1)

	service, err := m.node.Services().Register(ctx, m.node.Identity(), portName, queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for q := range queries {
		// allow local connections only
		if q.Source() != services.SourceLocal {
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
