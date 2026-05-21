package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

func (client *Client) Purge(ctx *astral.Context, repo string) (<-chan *astral.ObjectID, *error) {
	ch, err := client.queryCh(ctx, objects.MethodPurge, query.Args{
		"repo": repo,
	})
	if err != nil {
		return nil, &err
	}

	out := make(chan *astral.ObjectID)
	errPtr := new(error)

	go func() {
		defer close(out)
		defer ch.Close()

		*errPtr = ch.Switch(
			func(id *astral.ObjectID) error {
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

func Purge(ctx *astral.Context, repo string) (<-chan *astral.ObjectID, *error) {
	return Default().Purge(ctx, repo)
}
