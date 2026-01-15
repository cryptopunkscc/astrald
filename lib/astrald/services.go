package astrald

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/services"
)

type ServicesClient struct {
	c        *Client
	targetID *astral.Identity
}

func NewServicesClient(c *Client, targetID *astral.Identity) *ServicesClient {
	return &ServicesClient{c: c, targetID: targetID}
}

var defaultServicesClient *ServicesClient

func Services() *ServicesClient {
	if defaultServicesClient == nil {
		defaultServicesClient = NewServicesClient(DefaultClient(), nil)
	}
	return defaultServicesClient
}

func (client *ServicesClient) Discover(ctx *astral.Context, follow bool) (<-chan *services.Update, error) {
	ch, err := client.queryCh(ctx, "services.discover", query.Args{
		"follow": follow,
	})
	if err != nil {
		return nil, err
	}

	var out = make(chan *services.Update)

	go func() {
		defer ch.Close()
		defer close(out)
		for {
			obj, err := ch.Receive()
			if err != nil {
				return
			}
			switch obj := obj.(type) {
			case *services.Update:
				out <- obj
			case *astral.EOS:
				out <- nil
			default:
				return
			}
		}
	}()

	return out, nil
}

func (client *ServicesClient) queryCh(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return client.c.WithTarget(client.targetID).QueryChannel(ctx, method, args, cfg...)
}
