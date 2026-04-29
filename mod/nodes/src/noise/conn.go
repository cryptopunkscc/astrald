package noise

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/brontide"
)

var _ astral.Conn = &Conn{}

// Conn is an authenticated and encrypted connection produced by the Noise XK handshake.
type Conn struct {
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

func (conn *Conn) LocalIdentity() *astral.Identity {
	return astral.IdentityFromPubKey(conn.brontide.LocalPub())
}

func (conn *Conn) RemoteIdentity() *astral.Identity {
	return astral.IdentityFromPubKey(conn.brontide.RemotePub())
}
