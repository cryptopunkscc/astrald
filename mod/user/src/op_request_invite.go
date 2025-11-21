package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opRequestInviteArgs struct {
	Target string
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpRequestInvite(ctx *astral.Context, q shell.Query, args opRequestInviteArgs) (err error) {
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	ac := mod.ActiveContract()

	if args.Target == "" {
		if ac == nil {
			// We dont have an active contract to invite
			return q.RejectWithCode(2)
		}

		ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
		defer ch.Close()

		target := q.Caller()
		joinAllowed := mod.GetSwarmJoinRequestPolicy()(target)
		if !joinAllowed {
			return ch.Write(astral.NewError(user.ErrRequestDeclined.Error()))
		}

		signedContract, err := mod.ExchangeAndSignNodeContract(ctx, target, ac.UserID)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		err = mod.SaveSignedNodeContract(signedContract)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return ch.Write(signedContract)
	}

	if ac != nil {
		// We already have an active contract
		return q.RejectWithCode(2)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	requestInvitationQuery := query.New(mod.node.Identity(), target, user.OpRequestInvite, &opRequestInviteArgs{})
	requestCh, err := query.RouteChan(ctx, mod.node, requestInvitationQuery)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	obj, err := requestCh.Read()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	signedContract, ok := obj.(*user.SignedNodeContract)
	if !ok || signedContract == nil {
		return ch.Write(astral.NewError("invalid object when reading response to request invite"))
	}

	return ch.Write(signedContract)
}
