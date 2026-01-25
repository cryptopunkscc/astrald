package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"

	_ "github.com/cryptopunkscc/astrald/mod/allpub"
)

func (client *Client) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, *error) {
	ch, err := client.queryCh(ctx, "objects.describe", query.Args{
		"id": objectID.String(),
	})
	if err != nil {
		return nil, &err
	}

	out := make(chan *objects.DescribeResult)
	errPtr := new(error)

	go func() {
		defer close(out)
		defer ch.Close()

		*errPtr = ch.Switch(channel.Chan(out), channel.StopOnEOS, channel.PassErrors, channel.WithContext(ctx))
	}()

	return out, errPtr
}

func Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, *error) {
	return Default().Describe(ctx, objectID)
}
