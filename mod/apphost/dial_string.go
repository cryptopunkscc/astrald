package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
)

func (c *Server) handleDialString(ctx context.Context) error {
	var (
		err      error
		identity id.Identity
		nodeName string
		query    string
	)

	c.Decode("[c]c [c]c", &nodeName, &query)

	if nodeName == "" {
		identity = c.node.Identity()
	} else {
		identity, err = c.node.Contacts.ResolveIdentity(nodeName)
		if err != nil {
			return c.closeWithError(err)
		}
	}

	//TODO: Emit an event for logging?
	//log.Printf("(apphost) [%s] -> %s\n", c.node.Contacts.DisplayName(identity), query)

	conn, err := c.node.Query(ctx, identity, query)
	if err != nil {
		return c.closeWithError(err)
	}

	c.Encode("c", proto.ResponseOK)

	return join(ctx, c.conn, conn)
}
