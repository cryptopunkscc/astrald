package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opInviteArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpInvite(ctx *astral.Context, q *ops.Query, args opInviteArgs) (err error) {
	ac := mod.ActiveContract()
	if ac != nil {
		// We already have an active contract
		return q.RejectWithCode(2)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	obj, err := ch.Receive()
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	contract, ok := obj.(*user.NodeContract)
	if !ok || contract == nil {
		return ch.Send(user.ErrInvalidContract)
	}

	invitationAccepted := mod.GetSwarmInvitePolicy()(q.Caller(), *contract)
	if !invitationAccepted {
		return ch.Send(user.ErrInvitationDeclined)
	}

	if contract.UserID == nil {
		return ch.Send(user.ErrInvalidContract)
	}

	if !contract.NodeID.IsEqual(mod.node.Identity()) {
		return ch.Send(user.ErrInvalidContract)
	}

	if !contract.ExpiresAt.Time().After(time.Now().Add(minimalContractLength)) {
		return ch.Send(user.ErrInvalidContract)
	}

	signed := &user.SignedNodeContract{
		NodeContract: contract,
	}

	signed.NodeSig, err = mod.Keys.SignASN1(mod.ctx.Identity(), signed.Hash())
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = ch.Send(&signed.NodeSig)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	obj, err = ch.Receive()
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	userSig, ok := obj.(*astral.Bytes8)
	if !ok || userSig == nil {
		return ch.Send(user.ErrContractInvalidSignature)
	}

	signed.UserSig = *userSig
	err = mod.SaveSignedNodeContract(signed)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = mod.SetActiveContract(signed)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return nil
}
