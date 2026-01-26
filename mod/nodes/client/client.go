package nodes

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

func (client *Client) Relay(ctx *astral.Context, nonce astral.Nonce, caller *astral.Identity, target *astral.Identity) (err error) {
	ch, err := client.queryCh(ctx, "nodes.relay", query.Args{
		"nonce":      nonce,
		"set_caller": caller,
		"set_target": target,
	})
	if err != nil {
		return err
	}

	response, err := ch.Receive()
	if err != nil {
		return err
	}

	switch response := response.(type) {
	case *astral.ErrorMessage:
		return response

	case *astral.Ack:
		return nil

	default:
		return astral.NewErrUnexpectedObject(response)
	}
}

func (client *Client) queryCh(ctx *astral.Context, method string, args any) (*channel.Channel, error) {
	return client.astral.WithTarget(client.targetID).QueryChannel(ctx, method, args)
}
