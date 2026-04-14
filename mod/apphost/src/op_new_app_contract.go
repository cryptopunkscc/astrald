package apphost

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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

	permits := []*auth.Permit{
		{Action: astral.String8(nodes.RelayForAction{}.ObjectType())},
		{Action: astral.String8(apphost.HostForAction{}.ObjectType())},
	}

	return ch.Send(&auth.Contract{
		Issuer:    args.ID,
		Subject:   mod.node.Identity(),
		Permits:   astral.WrapSlice(&permits),
		ExpiresAt: astral.Time(time.Now().Add(time.Duration(args.Duration))),
	})
}
