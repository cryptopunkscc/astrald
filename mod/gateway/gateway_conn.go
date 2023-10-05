package gateway

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
)

var _ net.SecureConn = &GatewayConn{}

type GatewayConn struct {
	net.SecureConn
	localEndpoint  gw.Endpoint
	remoteEndpoint gw.Endpoint
}

func NewGatewayConn(conn net.SecureConn, localIdentity id.Identity, remoteIdentity id.Identity) *GatewayConn {
	c := &GatewayConn{
		SecureConn:     conn,
		localEndpoint:  gw.NewEndpoint(localIdentity, localIdentity),
		remoteEndpoint: gw.NewEndpoint(remoteIdentity, localIdentity),
	}
	return c
}

func (conn GatewayConn) LocalEndpoint() net.Endpoint {
	return conn.localEndpoint
}

func (conn GatewayConn) RemoteEndpoint() net.Endpoint {
	return conn.remoteEndpoint
}

func (GatewayConn) Network() string {
	return gw.DriverName
}
