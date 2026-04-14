package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opIndexArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpIndex(ctx *astral.Context, q *ops.Query, args opIndexArgs) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	sc, err := objects.Load[*auth.SignedContract](ctx, mod.Objects.ReadDefault(), args.ID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if err = mod.Auth.IndexContract(ctx, sc); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
