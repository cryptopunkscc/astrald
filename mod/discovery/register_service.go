package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
)

var _ tasks.Runner = &RegisterService{}

type RegisterService struct {
	*Module
}

const registerServiceName = "services.register"

func (m *RegisterService) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var workers = RunQueryWorkers(ctx, m.handleQuery, 2)
	var service, err = m.node.Services().Register(ctx, m.node.Identity(), registerServiceName, workers.Enqueue)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil
}

func (m *RegisterService) handleQuery(_ context.Context, query *services.Query) (err error) {
	conn, err := query.Accept()
	if err != nil {
		return err
	}

	return cslq.Invoke(conn, func(msg rpc.MsgRegister) error {
		var session = rpc.New(conn)
		var remoteID = conn.RemoteIdentity()

		source := &ServiceSource{
			services: m.node.Services(),
			service:  msg.Service,
		}

		// check if the caller is also the owner of the service
		service, err := m.node.Services().Find(msg.Service)
		if err != nil {
			defer conn.Close()
			return session.EncodeErr(rpc.ErrRegistrationFailed)
		} else {
			if !service.Identity().IsEqual(remoteID) {
				defer conn.Close()
				return session.EncodeErr(rpc.ErrRegistrationFailed)
			}
		}

		m.AddSource(source, conn.RemoteIdentity())

		go func() {
			<-service.Done()
			conn.Close()
		}()

		go func() {
			io.Copy(streams.NilWriter{}, conn)
			m.RemoveSource(source)
		}()

		return session.EncodeErr(nil)
	})
}
