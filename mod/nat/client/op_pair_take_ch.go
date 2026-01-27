package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairTakeCh(ctx *astral.Context, pair astral.Nonce, initiate bool) (*channel.Channel, error) {
	return client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodPairTake, query.Args{
		"pair":     pair,
		"initiate": initiate,
	})
}
