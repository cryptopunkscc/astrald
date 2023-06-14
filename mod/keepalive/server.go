package keepalive

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/query"
)

type Server struct {
	*Module
}

func NewServer(module *Module) *Server {
	return &Server{Module: module}
}

func (server *Server) Run(ctx context.Context) error {
	service, err := server.node.Services().Register(ctx, server.node.Identity(), serviceName, server)
	if err != nil {
		return err
	}

	<-service.Done()
	return nil
}

func (server *Server) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if linker, ok := remoteWriter.(query.Linker); ok {
		if l, ok := linker.Link().(*link.Link); ok {
			return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
				l.Idle().SetTimeout(0)

				server.log.Log("timeout disabled for %s over %s",
					l.RemoteIdentity(),
					l.Network(),
				)
			})
		}
	}

	return nil, link.ErrRejected
}
