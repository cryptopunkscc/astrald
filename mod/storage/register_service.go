package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage/proto"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &RegisterService{}

type RegisterService struct {
	*Module
}

func (m *RegisterService) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(4)
	service, err := m.node.Services().Register(ctx, m.node.Identity(), "storage.register", queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		conn, err := query.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := m.handle(ctx, conn); err != nil {
				m.log.Errorv(0, "register(): %s", err)
			}
		}()
	}

	return nil
}

func (m *RegisterService) handle(ctx context.Context, conn *services.Conn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg proto.MsgRegisterSource) error {
		var stream = proto.New(conn)

		source := &Source{
			Service:  msg.Service,
			Identity: conn.RemoteIdentity(),
		}

		m.AddSource(source)

		return stream.WriteError(nil)
	})
}
