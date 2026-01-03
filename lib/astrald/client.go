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

func NewClient(router Router) *Client {
	return &Client{Router: router}
}

func (client *Client) Query(ctx *astral.Context, method string, args any) (_ astral.Conn, err error) {
	return client.RouteQuery(ctx, query.New(client.GuestID(), client.targetID, method, args))
}

func (client *Client) QueryChannel(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	conn, err := client.Query(ctx, method, args)
	if err != nil {
		return nil, err
	}

	return channel.New(conn, cfg...), nil
}

func (client *Client) WithTarget(identity *astral.Identity) *Client {
	return &Client{Router: client.Router, targetID: identity}
}

func (client *Client) protocol() string {
	return "tcp"
}

func DefaultClient() *Client {
	if defaultClient == nil {
		defaultClient = NewClient(apphost.DefaultRouter())
	}
	return defaultClient
}

func SetDefaultClient(client *Client) {
	defaultClient = client
}

func RouteQuery(ctx *astral.Context, query *astral.Query) (astral.Conn, error) {
	return DefaultClient().RouteQuery(ctx, query)
}

func Query(ctx *astral.Context, method string, args any) (astral.Conn, error) {
	return DefaultClient().Query(ctx, method, args)
}

func QueryChannel(ctx *astral.Context, method string, args any, cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	return DefaultClient().QueryChannel(ctx, method, args, cfg...)
}

func WithTarget(identity *astral.Identity) *Client {
	return DefaultClient().WithTarget(identity)
}

func GuestID() *astral.Identity {
	return DefaultClient().GuestID()
}

func HostID() *astral.Identity {
	return DefaultClient().HostID()
}
