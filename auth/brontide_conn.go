package auth

import (
	"github.com/cryptopunkscc/astrald/auth/brontide"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.SecureConn = &NoiseConn{}

// NoiseConn is a net.SecureConn authenticated and ecrypted via the noise_xk protocol
type NoiseConn struct {
	conn     net.Conn
	brontide *brontide.Conn
}

func (conn *NoiseConn) Read(p []byte) (n int, err error) {
	return conn.brontide.Read(p)
}

func (conn *NoiseConn) Write(p []byte) (n int, err error) {
	return conn.brontide.Write(p)
}

func (conn *NoiseConn) Close() error {
	return conn.brontide.Close()
}

func (conn *NoiseConn) Outbound() bool {
	return conn.conn.Outbound()
}

func (conn *NoiseConn) LocalEndpoint() net.Endpoint {
	return conn.conn.LocalEndpoint()
}

func (conn *NoiseConn) RemoteEndpoint() net.Endpoint {
	return conn.conn.RemoteEndpoint()
}

func (conn *NoiseConn) LocalIdentity() id.Identity {
	return id.PublicKey(conn.brontide.LocalPub())
}

func (conn *NoiseConn) RemoteIdentity() id.Identity {
	return id.PublicKey(conn.brontide.RemotePub())
}
