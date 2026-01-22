package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Client struct {
	node     astral.Node
	astral   *astrald.Client
	callerID *astral.Identity
	targetID *astral.Identity
}

var defaultClient *Client

func New(targetID *astral.Identity, a *astrald.Client) *Client {
	if a == nil {
		a = astrald.DefaultClient()
	}
	return &Client{astral: a, callerID: a.GuestID(), targetID: targetID}
}

func NewFromNode(node astral.Node, targetID *astral.Identity) *Client {
	return &Client{node: node, callerID: node.Identity(), targetID: targetID}
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

func (client *Client) WithTarget(targetID *astral.Identity) *Client {
	return &Client{
		node:     client.node,
		astral:   client.astral,
		callerID: client.callerID,
		targetID: targetID,
	}
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	q := query.New(client.callerID, client.targetID, method, args)
	var conn astral.Conn
	var err error
	if client.node != nil {
		conn, err = query.Route(ctx, client.node, q)
	} else {
		conn, err = client.astral.RouteQuery(ctx, q)
	}
	if err != nil {
		return nil, err
	}
	return channel.New(conn, cfg...), nil
}
