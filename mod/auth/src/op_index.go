package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type opIndexArgs struct {
	ID  *astral.ObjectID `query:"required"`
	In  string           `query:"optional"`
	Out string           `query:"optional"`
}

func (mod *Module) OpIndex(ctx *astral.Context, q *ops.Query, args opIndexArgs) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	object, err := mod.Objects.Load(ctx, mod.Objects.ReadDefault(), args.ID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	signed, ok := object.(*auth.SignedContract)
	if !ok {
		return ch.Send(auth.ErrInvalidContract)
	}

	err = mod.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
