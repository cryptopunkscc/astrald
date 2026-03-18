package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	gw "github.com/cryptopunkscc/astrald/mod/gateway"
)

func (c *Client) List(ctx *astral.Context) ([]*astral.Identity, error) {
	ch, err := c.queryCh(ctx, gw.MethodNodeList, query.Args{})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var list []*astral.Identity
	err = ch.Switch(
		channel.Collect(&list),
		channel.StopOnEOS,
		channel.PassErrors,
	)

	return list, err
}
