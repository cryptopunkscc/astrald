package gateway

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/astral"
)

var _ exonet.Conn = &Conn{}

type Conn struct {
	astral.Conn
	localEndpoint  *Endpoint
	remoteEndpoint *Endpoint
	outbound       bool
}

func newConn(conn astral.Conn, localEndpoint *Endpoint, remoteEndpoint *Endpoint, outbound bool) *Conn {
	c := &Conn{
		Conn:           conn,
		localEndpoint:  localEndpoint,
		remoteEndpoint: remoteEndpoint,
		outbound:       outbound,
	}
	return c
}

func (conn Conn) LocalEndpoint() exonet.Endpoint {
	return conn.localEndpoint
}

func (conn Conn) RemoteEndpoint() exonet.Endpoint {
	return conn.remoteEndpoint
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}

func (Conn) Network() string {
	return NetworkName
}
