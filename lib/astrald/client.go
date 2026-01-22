package astrald

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Client struct {
	Router
	targetID *astral.Identity
}

var defaultClient *Client

func New(router Router) *Client {
	return &Client{Router: router}
}

func Default() *Client {
	if defaultClient == nil {
		defaultClient = New(apphost.DefaultRouter())
	}
	return defaultClient
}

func SetDefault(client *Client) {
	defaultClient = client
}

func (client *Client) Query(ctx *astral.Context, method string, args any) (_ astral.Conn, err error) {
	return client.RouteQuery(ctx, query.New(client.GuestID(), client.targetID, method, args))
}

func Query(ctx *astral.Context, method string, args any) (astral.Conn, error) {
	return Default().Query(ctx, method, args)
}

func (client *Client) QueryChannel(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	conn, err := client.Query(ctx, method, args)
	if err != nil {
		return nil, err
	}

	return channel.New(conn, cfg...), nil
}

func QueryChannel(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return Default().QueryChannel(ctx, method, args, cfg...)
}

func (client *Client) WithTarget(identity *astral.Identity) *Client {
	return &Client{Router: client.Router, targetID: identity}
}

func WithTarget(identity *astral.Identity) *Client {
	return Default().WithTarget(identity)
}

func (client *Client) Protocol() string {
	return "tcp"
}

func GuestID() *astral.Identity {
	return Default().GuestID()
}

func HostID() *astral.Identity {
	return Default().HostID()
}

func RouteQuery(ctx *astral.Context, query *astral.Query) (astral.Conn, error) {
	return Default().RouteQuery(ctx, query)
}
