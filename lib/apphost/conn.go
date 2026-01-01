package apphost

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ net.Conn = &Conn{}

type Conn struct {
	net.Conn
	remoteID *astral.Identity
	localID  *astral.Identity
	query    string
	nonce    astral.Nonce
}

func NewConn(conn net.Conn, remoteID *astral.Identity, localID *astral.Identity, query string, nonce astral.Nonce) *Conn {
	return &Conn{Conn: conn, remoteID: remoteID, localID: localID, query: query, nonce: nonce}
}

func (conn Conn) RemoteIdentity() *astral.Identity {
	return conn.remoteID
}

func (conn Conn) LocalIdentity() *astral.Identity {
	return conn.localID
}

func (conn Conn) RemoteAddr() net.Addr {
	return Addr{address: conn.remoteID.String()}
}

func (conn Conn) Query() string {
	return conn.query
}

func (conn Conn) ID() astral.Nonce {
	return conn.nonce
}
