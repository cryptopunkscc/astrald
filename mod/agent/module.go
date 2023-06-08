package agent

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
)

const portName = "sys.agent"

type Module struct {
	node node.Node
	log  *log.Logger
}

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
		if q.Origin() != services.OriginLocal {
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
			log:  m.log,
		}

		go func() {
			if err := s.Run(ctx); err != nil {
				m.log.Error("serve error: %s", err)
			}
		}()
	}

	return nil
}
