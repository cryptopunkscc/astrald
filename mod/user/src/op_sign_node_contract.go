package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opSignNodeContractArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSignNodeContract(ctx *astral.Context, query *routing.IncomingQuery, args opSignNodeContractArgs) (err error) {
	ch := query.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Switch(func(contract *user.NodeContract) error {
		signer, err := mod.SignNodeContract(ctx, contract)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(signer)
	})
}
