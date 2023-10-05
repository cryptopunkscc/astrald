package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/streams"
)

type GatewayService struct {
	*Module
}

func (srv *GatewayService) Run(ctx context.Context) error {
	service, err := srv.node.Services().Register(ctx, srv.node.Identity(), gw.ServiceName, srv)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil

}

func (srv *GatewayService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer debug.SaveLog(func(p any) {
			srv.log.Error("gateway panicked: %v", p)
		})

		srv.handleQuery(conn)
	})
}

func (srv *GatewayService) handleQuery(conn net.SecureConn) error {
	var err error
	var c = cslq.NewEndec(conn)
	var cookie string

	err = c.Decodef("[c]c", &cookie)
	if err != nil {
		return err
	}

	nodeID, err := id.ParsePublicKeyHex(cookie)
	if err != nil {
		srv.log.Errorv(2, "invalid request from %v: malformed target identity", conn.RemoteIdentity())
		c.Encodef("c", false)
		conn.Close()
		return err
	}

	out, err := net.Route(srv.ctx, srv.node.Network(), net.NewQuery(srv.node.Identity(), nodeID, ConnectServiceName))
	if err != nil {
		conn.Close()
		return err
	}

	c.Encodef("c", true)

	l, r, err := streams.Join(conn, out)

	srv.log.Logv(1, "conn for %s done (bytes read %d written %d)", nodeID, l, r)

	return err
}
