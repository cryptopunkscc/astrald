package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

type Client struct {
	astral   *astrald.Client
	targetID *astral.Identity
}

// New creates a Client targeting targetID; falls back to astrald.Default() when a is nil.
func New(targetID *astral.Identity, a *astrald.Client) *Client {
	if a == nil {
		a = astrald.Default()
	}
	return &Client{astral: a, targetID: targetID}
}

// WithTarget returns a new Client bound to target, sharing the same underlying astral client.
func (client *Client) WithTarget(target *astral.Identity) *Client {
	return &Client{astral: client.astral, targetID: target}
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args, cfg...)
}
