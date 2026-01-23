package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

type Client struct {
	astral   *astrald.Client
	targetID *astral.Identity
}

var defaultClient *Client

func New(targetID *astral.Identity, a *astrald.Client) *Client {
	if a == nil {
		a = astrald.Default()
	}
	return &Client{astral: a, targetID: targetID}
}

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, nil)
	}
	return defaultClient
}

func SetDefault(client *Client) {
	defaultClient = client
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.astral.QueryChannel(ctx, method, args, cfg...)
}
