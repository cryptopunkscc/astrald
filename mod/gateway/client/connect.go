package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	gw "github.com/cryptopunkscc/astrald/mod/gateway"
)

func (c *Client) Connect(ctx *astral.Context, target *astral.Identity) (*gw.Socket, error) {
	ch, err := c.queryCh(ctx, gw.MethodNodeConnect, query.Args{"target": target.String()})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var socket *gw.Socket
	err = ch.Switch(
		channel.Expect(&socket),
		channel.PassErrors,
	)

	return socket, err
}
