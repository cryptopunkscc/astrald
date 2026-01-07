package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opClaimArgs struct {
	Target   string
	StartsAt *astral.Time    `query:"optional"`
	StartsIn astral.Duration `query:"optional"`
	In       string          `query:"optional"`
	Out      string          `query:"optional"`
}

func (mod *Module) OpClaim(ctx *astral.Context, q shell.Query, args opClaimArgs) (err error) {
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	var startsAt astral.Time
	switch {
	case args.StartsAt != nil:
		startsAt = *args.StartsAt
	case args.StartsIn != 0:
		startsAt = astral.Time(astral.Now().Time().Add(args.StartsIn.Duration()))
	default:
		startsAt = astral.Now() // default immediate start
	}

	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	if !q.Caller().IsEqual(ac.UserID) {
		return q.RejectWithCode(3)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	signedContract, err := mod.ExchangeAndSignNodeContract(ctx, target, ac.UserID, startsAt)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = mod.SaveSignedNodeContract(signedContract)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(signedContract)
}
