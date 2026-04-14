package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
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

	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// receive the contract to sign
	var contract *auth.Contract
	err = ch.Switch(channel.Expect(&contract))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// check contract viability
	switch {
	case contract.Subject.IsZero():
		return ch.Send(auth.ErrInvalidContract)
	case !contract.Subject.IsEqual(mod.node.Identity()):
		return ch.Send(auth.ErrInvalidContract)
	case contract.ExpiresAt.Time().Before(time.Now().Add(minimalContractLength)):
		return ch.Send(auth.ErrInvalidContract)
	}

	// wait for user approval
	approved := mod.GetSwarmInvitePolicy()(q.Caller(), contract)
	if !approved {
		return ch.Send(user.ErrInvitationDeclined)
	}

	// sign the contract as subject (node)
	subjectSig, err := mod.Auth.SignSubject(ctx, contract)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// send signature back
	err = ch.Send(subjectSig)
	if err != nil {
		return
	}

	// expect issuer signature
	var issuerSig *crypto.Signature
	err = ch.Switch(channel.Expect(&issuerSig))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// assemble the signed contract
	signed := &auth.SignedContract{
		Contract:  contract,
		IssuerSig: issuerSig,
		SubjecSig: subjectSig,
	}

	// final signature verification
	err = mod.Auth.VerifyContract(signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.Auth.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// store the signed contract
	_, err = mod.Objects.Store(ctx, mod.Objects.WriteDefault(), signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// set the contract as the active contract
	err = mod.SetActiveContract(signed)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return nil
}
