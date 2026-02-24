package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/kcp"
)

func (client *Client) SetEndpointLocalPort(ctx *astral.Context, endpoint kcp.Endpoint, localPort astral.Uint16, replace bool) error {
	ch, err := client.queryCh(ctx, kcp.MethodSetEndpointLocalPort, query.Args{
		"endpoint":   endpoint.Address(),
		"local_port": localPort,
		"replace":    replace,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(
		channel.ExpectAck,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)
}
