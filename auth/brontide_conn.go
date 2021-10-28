package auth

import (
	"github.com/cryptopunkscc/astrald/auth/brontide"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

type brontideConn struct {
	netConn        infra.Conn
	bConn          *brontide.Conn
	remoteIdentity *id.Identity
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
	return conn.netConn.Outbound()
}

func (conn *brontideConn) LocalAddr() infra.Addr {
	return conn.netConn.LocalAddr()
}

func (conn *brontideConn) RemoteAddr() infra.Addr {
	return conn.netConn.RemoteAddr()
}

func (conn *brontideConn) LocalIdentity() id.Identity {
	return id.PublicKey(conn.bConn.LocalPub())
}

func (conn *brontideConn) RemoteIdentity() id.Identity {
	return id.PublicKey(conn.bConn.RemotePub())
}
