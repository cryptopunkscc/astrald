package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
)

func (c *Server) handleInfo(_ context.Context) error {
	return c.Encode("c v", proto.ResponseOK, c.node.Identity())
}
