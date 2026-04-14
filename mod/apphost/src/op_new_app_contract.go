package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type opNewAppContractArgs struct {
	ID       *astral.Identity
	Duration astral.Duration `query:"optional"`
	Out      string          `query:"optional"`
}

func (mod *Module) OpNewAppContract(ctx *astral.Context, q *ops.Query, args opNewAppContractArgs) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if args.Duration == 0 {
		args.Duration = DefaultAppContractDuration
	}

	contract, err := apphost.NewAppContract(args.ID, mod.node.Identity(), args.Duration.Duration())
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(contract)
}
