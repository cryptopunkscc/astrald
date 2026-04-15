package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type opIndexArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpIndex(ctx *astral.Context, q *ops.Query, args opIndexArgs) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var signed *auth.SignedContract

	err := ch.Switch(
		func(objectID *astral.ObjectID) error {
			object, loadErr := mod.Objects.Load(ctx, mod.Objects.ReadDefault(), objectID)
			if loadErr != nil {
				return loadErr
			}
			var ok bool
			signed, ok = object.(*auth.SignedContract)
			if !ok {
				return auth.ErrInvalidContract
			}
			return channel.ErrBreak
		},
		func(sc *auth.SignedContract) error {
			signed = sc
			return channel.ErrBreak
		},
		channel.PassErrors,
	)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.EOS{})
}
