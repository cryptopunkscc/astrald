package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type opSignContractArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSignContract(ctx *astral.Context, q *routing.IncomingQuery, args opSignContractArgs) error {
	ch := q.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Switch(func(c *auth.Contract) error {
		signed := &auth.SignedContract{Contract: c}
		err := mod.SignContract(ctx, signed)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(signed)
	})
}
