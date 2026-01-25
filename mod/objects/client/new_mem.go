package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) NewMem(ctx *astral.Context, name string, size int64) error {
	// send the query
	ch, err := client.queryCh(ctx, "objects.new_mem", query.Args{
		"name": name,
		"size": size,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	// wait for ack
	return ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
}
