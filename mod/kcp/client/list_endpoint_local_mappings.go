package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/kcp"
)

func (client *Client) ListEndpointLocalMappings(ctx *astral.Context) ([]*kcp.EndpointLocalMapping, error) {
	ch, err := client.queryCh(ctx, kcp.MethodListEndpointsLocalMappings, nil)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var mappings []*kcp.EndpointLocalMapping

	err = ch.Switch(
		channel.Collect(&mappings),
		channel.StopOnEOS,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)

	return mappings, err
}
