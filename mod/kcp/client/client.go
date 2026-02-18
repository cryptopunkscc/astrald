package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

type Client struct {
	astral *astrald.Client
}

func New(a *astrald.Client) *Client {
	if a == nil {
		a = astrald.Default()
	}
	return &Client{astral: a}
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.astral.QueryChannel(ctx, method, args, cfg...)
}
