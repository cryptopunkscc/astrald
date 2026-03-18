package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	gw "github.com/cryptopunkscc/astrald/mod/gateway"
)

func (c *Client) Bind(ctx *astral.Context, visibility gw.Visibility) (*gw.Socket, error) {
	ch, err := c.queryCh(ctx, gw.MethodNodeRegister, query.Args{"visibility": string(visibility)})
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
