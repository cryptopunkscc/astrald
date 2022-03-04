package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
)

func (c *Server) handleDialKey(ctx context.Context) error {
	var (
		identity id.Identity
		query    string
	)

	c.Decode("v[c]c", &identity, &query)

	if identity.IsZero() {
		identity = c.node.Identity()
	}

	conn, err := c.node.Query(ctx, identity, query)
	if err != nil {
		c.Encode("c", proto.ResponseRejected)
		c.conn.Close()
		return err
	}

	c.Encode("c", proto.ResponseOK)

	return join(ctx, c.conn, conn)
}
