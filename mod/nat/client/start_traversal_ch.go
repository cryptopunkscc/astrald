package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) StartTraversalCh(ctx *astral.Context, out string) (*channel.Channel, error) {
	return client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodStartNatTraversal, query.Args{
		"out": out,
	})
}
