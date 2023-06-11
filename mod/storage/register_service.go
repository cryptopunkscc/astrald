package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage/rpc"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &RegisterService{}

const RegisterServiceName = "storage.register"

type RegisterService struct {
	*Module
}

func (m *RegisterService) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(4)
	service, err := m.node.Services().Register(ctx, m.node.Identity(), RegisterServiceName, queries.Push)
	if err != nil {
		m.log.Error("cannot register service %s: %s", RegisterServiceName, err)
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		query := query
		go func() {
			if err := m.handle(ctx, query); err != nil {
				m.log.Errorv(0, "RegisterService.handle(): %s", err)
			}
		}()
	}

	return nil
}

func (m *RegisterService) handle(ctx context.Context, query *services.Query) error {
	if !m.IsProvider(query.RemoteIdentity()) {
		m.log.Errorv(2, "register_provider: %v is not a provider, rejecting...", query.RemoteIdentity())
		query.Reject()
		return nil
	}

	conn, err := query.Accept()
	if err != nil {
		return err
	}

	defer conn.Close()
	return cslq.Invoke(conn, func(msg rpc.MsgRegisterSource) error {
		var session = rpc.New(conn)

		source := &DataSource{
			Service:  msg.Service,
			Identity: conn.RemoteIdentity(),
		}

		m.AddDataSource(source)

		return session.EncodeErr(nil)
	})
}
