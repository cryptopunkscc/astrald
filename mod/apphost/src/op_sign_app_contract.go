package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type opSignAppContractArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSignAppContract(ctx *astral.Context, q *ops.Query, args opSignAppContractArgs) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Switch(func(c *auth.Contract) error {
		signed, err := mod.Auth.SignContract(ctx, c)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		_, err = mod.Objects.Store(ctx, mod.Objects.WriteDefault(), signed)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		err = mod.Auth.IndexContract(ctx, signed)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		mod.log.Logv(1, "signed app contract (%v)", signed.Issuer)

		return ch.Send(signed)
	})
}
