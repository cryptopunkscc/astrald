package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opRequestInviteArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRequestInvite(ctx *astral.Context, q *ops.Query, args opRequestInviteArgs) (err error) {
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	ac := mod.ActiveContract()

	if ac == nil {
		// We dont have an active contract to invite
		return q.RejectWithCode(2)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	target := q.Caller()
	joinAllowed := mod.GetSwarmJoinRequestPolicy()(target)
	if !joinAllowed {
		return ch.Send(user.ErrRequestDeclined)
	}

	signed, err := mod.InviteNode(ctx, target)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	_, err = mod.IndexSignedNodeContract(signed)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(signed)
}
