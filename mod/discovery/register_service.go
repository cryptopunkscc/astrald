package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
)

var _ tasks.Runner = &RegisterService{}

type RegisterService struct {
	*Module
}

const registerServiceName = "services.register"

func (srv *RegisterService) Run(ctx context.Context) error {
	var service, err = srv.node.Services().Register(ctx, srv.node.Identity(), registerServiceName, srv)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil
}

func (srv *RegisterService) RouteQuery(ctx context.Context, q query.Query, swc net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return query.Accept(q, swc, func(conn net.SecureConn) {
		cslq.Invoke(conn, func(msg rpc.MsgRegister) error {
			return srv.serveRegister(conn, msg)
		})
	})
}

func (srv *RegisterService) serveRegister(conn net.SecureConn, msg rpc.MsgRegister) (err error) {
	var session = rpc.New(conn)
	var remoteID = conn.RemoteIdentity()

	source := &ServiceSource{
		services: srv.node.Services(),
		identity: remoteID,
		service:  msg.Service,
	}

	// check if the caller is also the owner of the service
	service, err := srv.node.Services().Find(msg.Service)
	if err != nil {
		defer conn.Close()
		return session.EncodeErr(rpc.ErrRegistrationFailed)
	} else {
		if !service.Identity().IsEqual(remoteID) {
			defer conn.Close()
			return session.EncodeErr(rpc.ErrRegistrationFailed)
		}
	}

	srv.AddSource(source, conn.RemoteIdentity())

	go func() {
		<-service.Done()
		conn.Close()
	}()

	go func() {
		io.Copy(streams.NilWriter{}, conn)
		srv.RemoveSource(source)
	}()

	return session.EncodeErr(nil)
}
