package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairLock(ctx *astral.Context, pair astral.Nonce) error {
	ch, err := client.queryCh(ctx, nat.MethodPairLock, query.Args{
		"pair": pair,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Switch(
		channel.ExpectAck,
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)
}
