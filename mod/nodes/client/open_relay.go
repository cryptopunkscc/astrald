package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	nodesmod "github.com/cryptopunkscc/astrald/mod/nodes"
)

func (client *Client) OpenRelay(ctx *astral.Context) (*channel.Channel, error) {
	ch, err := client.queryCh(ctx, nodesmod.MethodNodeOpenRelay, nil)
	if err != nil {
		return nil, err
	}

	return ch, nil
}
