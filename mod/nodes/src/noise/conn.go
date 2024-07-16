package noise

import (
	"github.com/cryptopunkscc/astrald/brontide"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Conn = &Conn{}

// Conn is a net.SecureConn authenticated and ecrypted via the noise_xk protocol
type Conn struct {
	conn     exonet.Conn
	brontide *brontide.Conn
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	return conn.brontide.Read(p)
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	return conn.brontide.Write(p)
}

func (conn *Conn) Close() error {
	return conn.brontide.Close()
}

func (conn *Conn) Outbound() bool {
	return conn.conn.Outbound()
}

func (conn *Conn) LocalEndpoint() exonet.Endpoint {
	return conn.conn.LocalEndpoint()
}

func (conn *Conn) RemoteEndpoint() exonet.Endpoint {
	return conn.conn.RemoteEndpoint()
}

func (conn *Conn) LocalIdentity() id.Identity {
	return id.PublicKey(conn.brontide.LocalPub())
}

func (conn *Conn) RemoteIdentity() id.Identity {
	return id.PublicKey(conn.brontide.RemotePub())
}
