package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

const ConnectServiceName = "connect"

type ConnectService struct {
	*Module
}

func (srv *ConnectService) Run(ctx context.Context) error {
	s, err := srv.node.Services().Register(ctx, srv.node.Identity(), ConnectServiceName, srv)
	if err != nil {
		return err
	}

	<-s.Done()
	return nil
}

func (srv *ConnectService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		l, err := link.Accept(ctx, NewGatewayConn(conn, query.Target(), query.Caller()), srv.node.Identity())
		if err != nil {
			return
		}

		err = srv.node.Network().AddLink(l)
		if err != nil {
			l.Close()
		}
	})

}
