package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
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
	var contract *user.NodeContract
	err = ch.Switch(channel.Expect(&contract))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// check contract viability
	switch {
	case contract.NodeID.IsZero():
		return ch.Send(user.ErrInvalidContract)
	case !contract.NodeID.IsEqual(mod.node.Identity()):
		return ch.Send(user.ErrInvalidContract)
	case !contract.ActiveAt(time.Now().Add(minimalContractLength)):
		return ch.Send(user.ErrInvalidContract)
	}

	// wait for user approval
	approved := mod.GetSwarmInvitePolicy()(q.Caller(), *contract)
	if !approved {
		return ch.Send(user.ErrInvitationDeclined)
	}

	// get the signer for the node
	signer, err := mod.Crypto.HashSigner(secp256k1.FromIdentity(mod.node.Identity()), crypto.SchemeASN1)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// sign the contract
	nodeSig, err := signer.SignHash(ctx, contract.ContractHash())
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// send signature back
	err = ch.Send(nodeSig)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	var userSig *crypto.Signature
	err = ch.Switch(channel.Expect(&userSig))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// assemble the signed contract
	signed := &user.SignedNodeContract{
		NodeContract: contract,
		UserSig:      userSig,
		NodeSig:      nodeSig,
	}

	// final signature verification
	err = mod.VerifySignedNodeContract(signed)
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
