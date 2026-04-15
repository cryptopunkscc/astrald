package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opInviteArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpInvite(ctx *astral.Context, q *routing.IncomingQuery, args opInviteArgs) (err error) {
	ac := mod.ActiveContract()
	if ac != nil {
		// We already have an active contract
		return q.RejectWithCode(2)
	}

	ch := q.Accept(channel.WithFormats(args.In, args.Out))
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

	approved := mod.GetSwarmInvitePolicy()(q.Caller(), contract)
	if !approved {
		return ch.Send(user.ErrInvitationDeclined)
	}

	var issuerSig *crypto.Signature
	err = ch.Switch(channel.Expect(&issuerSig))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	signed := &auth.SignedContract{Contract: contract, IssuerSig: issuerSig}
	if err = mod.Auth.VerifyIssuer(signed); err != nil {
		return ch.Send(astral.Err(err))
	}

	subjectSig, err := mod.Auth.SignSubject(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = ch.Send(subjectSig)
	if err != nil {
		return
	}

	err = mod.Auth.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	_, err = mod.Objects.Store(ctx, mod.Objects.WriteDefault(), signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.SetActiveContract(signed)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return nil
}
