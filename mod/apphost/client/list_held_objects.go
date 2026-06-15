package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

// ListHeldObjects streams held object IDs asynchronously; the error pointer is populated only after the channel closes.
func (client *Client) ListHeldObjects(ctx *astral.Context) (<-chan *astral.ObjectID, *error) {
	ch, err := client.queryCh(ctx, apphost.MethodListHeldObjects, nil)
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

func ListHeldObjects(ctx *astral.Context) (<-chan *astral.ObjectID, *error) {
	return Default().ListHeldObjects(ctx)
}
