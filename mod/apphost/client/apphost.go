package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Client struct {
	astral   *astrald.Client
	targetID *astral.Identity
}

var defaultClient *Client

func New(targetID *astral.Identity, client *astrald.Client) *Client {
	if client == nil {
		client = astrald.Default()
	}
	return &Client{astral: client, targetID: targetID}
}

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(nil, astrald.Default())
	}
	return defaultClient
}

// RegisterHandler registers a new handler for incoming queries.
func (client *Client) RegisterHandler(ctx *astral.Context) (*astrald.Listener, error) {
	l, err := astrald.NewListener(client.astral.Protocol(), astral.NewNonce())
	if err != nil {
		return nil, err
	}

	ch, err := client.queryCh(ctx, "apphost.register_handler", query.Args{
		"endpoint": l.String(),
		"token":    l.Token(),
	})
	if err != nil {
		l.Close()
		return nil, err
	}
	defer ch.Close()

	o, err := ch.Receive()
	switch o.(type) {
	case *astral.Ack:
		return l, nil

	case nil:
		l.Close()
		return nil, err

	default:
		return nil, astral.NewErrUnexpectedObject(o)
	}
}

func RegisterHandler(ctx *astral.Context) (*astrald.Listener, error) {
	return Default().RegisterHandler(ctx)
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args)
}
