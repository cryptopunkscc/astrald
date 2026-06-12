package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

// Client is an RPC client for the gateway module service.
// A nil targetID routes queries to the local node's gateway.
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

// Default returns the process-wide default Client, creating it on first call with nil identity and astrald.Default() as the client.
func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, nil)
	}
	return defaultClient
}

func SetDefault(client *Client) {
	defaultClient = client
}

func (c *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return c.astral.WithTarget(c.targetID).QueryChannel(ctx, method, args, cfg...)
}
