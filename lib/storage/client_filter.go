package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	jrpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"io"
)

func (c *Client) IdFilter() (filter id.Filter, closer io.Closer, err error) {
	closer = c.conn
	filter = func(identity id.Identity) (b bool) {
		if identity.IsEqual(id.Anyone) {
			return true
		}
		b, _ = jrpc.Query[bool](c.conn, "", identity)
		return b
	}
	return
}
