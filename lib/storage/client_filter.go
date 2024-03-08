package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	proto "github.com/cryptopunkscc/astrald/mod/storage/srv"
	"io"
)

func (c *Client) IdFilter() (filter id.Filter, closer io.Closer, err error) {
	conn, err := c.query(nil)
	if err != nil {
		return
	}
	closer = conn
	enc := proto.NewBinaryEncoder(conn)
	filter = func(identity id.Identity) (b bool) {
		if identity.PublicKey() == nil {
			return true
		}
		bytes := identity.PublicKey().SerializeCompressed()
		if err := enc.Encode(bytes); err != nil {
			return
		}
		if err := enc.Decode(&b); err != nil {
			return
		}
		return
	}
	return
}
