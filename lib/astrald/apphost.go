package astrald

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type AppHostClient struct {
	c        *Client
	targetID *astral.Identity
}

var defaultAppHostClient *AppHostClient

func NewAppHostClient(targetID *astral.Identity, client *Client) *AppHostClient {
	if client == nil {
		client = DefaultClient()
	}

	return &AppHostClient{c: client, targetID: targetID}
}

func AppHost() *AppHostClient {
	if defaultAppHostClient == nil {
		defaultAppHostClient = NewAppHostClient(nil, nil)
	}
	return defaultAppHostClient
}

// RegisterHandler registers a new handler for incoming queries.
func (client *AppHostClient) RegisterHandler(ctx *astral.Context) (*Listener, error) {
	l, err := NewListener(client.c.protocol(), astral.NewNonce())
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
		return nil, errors.New("unexpected response")
	}
}

func (client *AppHostClient) queryCh(ctx *astral.Context, method string, args any) (*channel.Channel, error) {
	return client.c.WithTarget(client.targetID).QueryChannel(ctx, method, args)
}
