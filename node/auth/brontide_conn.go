package auth

import (
	"github.com/cryptopunkscc/astrald/node/auth/brontide"
	"github.com/cryptopunkscc/astrald/node/net"
)

type brontideConn struct {
	netConn        net.Conn
	bConn          *brontide.Conn
	remoteIdentity Identity
}

func (conn *brontideConn) Read(p []byte) (n int, err error) {
	return conn.bConn.Read(p)
}

func (conn *brontideConn) Write(p []byte) (n int, err error) {
	return conn.bConn.Write(p)
}

func (conn *brontideConn) Close() error {
	return conn.bConn.Close()
}

func (conn *brontideConn) Outbound() bool {
	return false
}

func (conn *brontideConn) RemoteEndpoint() net.Endpoint {
	return net.Endpoint{}
}

func (conn *brontideConn) RemoteIdentity() Identity {
	return &ECIdentity{
		publicKey: conn.bConn.RemotePub(),
	}
}
