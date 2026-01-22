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

func New(targetID *astral.Identity, astral *astrald.Client) *Client {
	if astral == nil {
		astral = astrald.DefaultClient()
	}
	return &Client{astral: astral, targetID: targetID}
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
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args, cfg...)
}
