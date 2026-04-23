package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	nodesmod "github.com/cryptopunkscc/astrald/mod/nodes"
)

type MigrateSessionArgs struct {
	SessionID astral.Nonce
	StreamID  astral.Nonce
}

func (client *Client) MigrateSession(ctx *astral.Context, args MigrateSessionArgs) (*channel.Channel, error) {
	return client.queryCh(ctx, nodesmod.MethodMigrateSession, args)
}
