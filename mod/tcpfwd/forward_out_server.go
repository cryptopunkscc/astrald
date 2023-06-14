package tcpfwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/streams"
	_net "net"
)

type ForwardOutServer struct {
	*Module
	serviceName string
	target      string
}

func (server *ForwardOutServer) Run(ctx context.Context) error {
	s, err := server.node.Services().Register(ctx, server.node.Identity(), server.serviceName, server)
	if err != nil {
		return err
	}

	server.log.Logv(1, "forwarding %s to %s", server.serviceName, server.target)

	<-s.Done()
	return nil
}

func (server *ForwardOutServer) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	outConn, err := _net.Dial("tcp", server.target)
	if err != nil {
		server.log.Errorv(1, "error forwarding %s to %s: %s", server.serviceName, server.target, err)
		return nil, err
	}

	return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
		streams.Join(conn, outConn)
	})
}
