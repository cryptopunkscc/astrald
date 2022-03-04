package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

func (c *Server) handleGetNodeName(_ context.Context) error {
	var remoteID id.Identity

	if err := c.Decode("v", &remoteID); err != nil {
		return err
	}

	return c.Encode("[c]c", c.node.Contacts.DisplayName(remoteID))
}
