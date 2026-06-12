package crypto

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

// New builds a client targeting targetID; a nil astral falls back to astrald.Default.
func New(targetID *astral.Identity, astral *astrald.Client) *Client {
	if astral == nil {
		astral = astrald.Default()
	}
	return &Client{astral: astral, targetID: targetID}
}

// Default returns the lazily-initialized package-wide client with no target set.
func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, astrald.Default())
	}
	return defaultClient
}

func (client *Client) query(ctx *astral.Context, method string, args any) (astral.Conn, error) {
	return client.astral.WithTarget(client.targetID).Query(ctx, method, args)
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args, cfg...)
}
