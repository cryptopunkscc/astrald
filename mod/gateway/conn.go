package gateway

import (
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.SecureConn = &Conn{}

type Conn struct {
	net.SecureConn
	localEndpoint  Endpoint
	remoteEndpoint Endpoint
	outbound       bool
}

func newConn(conn net.SecureConn, localEndpoint Endpoint, remoteEndpoint Endpoint, outbound bool) *Conn {
	c := &Conn{
		SecureConn:     conn,
		localEndpoint:  localEndpoint,
		remoteEndpoint: remoteEndpoint,
		outbound:       outbound,
	}
	return c
}

func (conn Conn) LocalEndpoint() net.Endpoint {
	return conn.localEndpoint
}

func (conn Conn) RemoteEndpoint() net.Endpoint {
	return conn.remoteEndpoint
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}

func (Conn) Network() string {
	return NetworkName
}
