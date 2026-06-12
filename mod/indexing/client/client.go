package indexing

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

// Client is an RPC client for the indexing module. A nil targetID routes
// calls to the local node's indexing service.
type Client struct {
	astral   *astrald.Client
	targetID *astral.Identity
}

// New creates a Client targeting the given identity. If a is nil, astrald.Default()
// is used; if targetID is nil, calls are routed to the local node.
func New(targetID *astral.Identity, a *astrald.Client) *Client {
	if a == nil {
		a = astrald.Default()
	}
	return &Client{astral: a, targetID: targetID}
}

var defaultClient *Client

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, astrald.Default())
	}
	return defaultClient
}

func (c *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return c.astral.WithTarget(c.targetID).QueryChannel(ctx, method, args, cfg...)
}
