package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

func (client *Client) Find(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *astral.Identity, *error) {
	ch, err := client.queryCh(ctx, objects.MethodFind, query.Args{
		"id": objectID,
	})
	if err != nil {
		return nil, &err
	}

	out := make(chan *astral.Identity)
	errPtr := new(error)

	go func() {
		defer close(out)
		defer ch.Close()

		*errPtr = ch.Switch(
			func(id *astral.Identity) error {
				if id != nil && !id.IsZero() {
					return sig.Send(ctx, out, id)
				}
				return nil
			},
			channel.BreakOnEOS,
			channel.PassErrors,
			channel.WithContext(ctx),
		)
	}()

	return out, errPtr
}

func Find(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *astral.Identity, *error) {
	return Default().Find(ctx, objectID)
}
