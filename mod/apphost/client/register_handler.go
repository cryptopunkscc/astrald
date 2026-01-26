package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

// RegisterHandler registers a new handler for incoming queries. Protocol is tcp
func (client *Client) RegisterHandler(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	ch, err := client.queryCh(ctx, "apphost.register_handler", query.Args{
		"endpoint": endpoint,
		"token":    authToken,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(channel.ExpectAck, channel.PassErrors)
}

func RegisterHandler(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	return Default().RegisterHandler(ctx, endpoint, authToken)
}
