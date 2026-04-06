package apphost

import (
	"time"

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

	return ch.Send(&apphost.AppContract{
		AppID:     args.ID,
		HostID:    mod.node.Identity(),
		StartsAt:  astral.Time(time.Now()),
		ExpiresAt: astral.Time(time.Now().Add(time.Duration(args.Duration))),
	})
}
