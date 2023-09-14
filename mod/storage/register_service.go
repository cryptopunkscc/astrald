package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage/rpc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &RegisterService{}

const RegisterServiceName = "storage.register"

type RegisterService struct {
	*Module
}

func (service *RegisterService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), RegisterServiceName, service)
	if err != nil {
		service.log.Error("cannot register service %s: %s", RegisterServiceName, err)
		return err
	}

	<-s.Done()

	return nil
}

func (service *RegisterService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !service.IsProvider(query.Caller()) {
		service.log.Errorv(2, "register_provider: %v is not a provider, rejecting...", query.Caller())
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		service.handle(service.ctx, conn)
	})
}

func (service *RegisterService) handle(ctx context.Context, conn net.SecureConn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg rpc.MsgRegisterSource) error {
		var session = rpc.New(conn)

		source := &DataSource{
			Service:  msg.Service,
			Identity: conn.RemoteIdentity(),
		}

		service.AddDataSource(source)

		return session.EncodeErr(nil)
	})
}
