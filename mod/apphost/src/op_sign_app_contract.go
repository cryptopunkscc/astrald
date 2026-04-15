package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type opSignAppContractArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSignAppContract(ctx *astral.Context, q *routing.IncomingQuery, args opSignAppContractArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Switch(func(c *apphost.AppContract) error {
		signed, err := mod.SignAppContract(ctx, c)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		mod.log.Logv(1, "signed app contract (%v)", signed.AppID)

		return ch.Send(signed)
	})
}
