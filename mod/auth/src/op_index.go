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
	err := ch.Switch(channel.Expect(&signed), channel.PassErrors)
	if err != nil {
		return err
	}

	err = mod.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.EOS{})
}
