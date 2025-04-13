package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type opInviteArgs struct {
}

// OpInvite reads a contract request and writes a signed contract or an error.
func (mod *Module) OpInvite(ctx *astral.Context, q shell.Query, args opInviteArgs) (err error) {
	// reject if we have an active contract already
	ac := mod.ActiveContract()
	if ac != nil {
		return q.Reject()
	}

	conn := q.Accept()
	defer conn.Close()

	// read a NodeContract
	obj, _, err := mod.Objects.Blueprints().Read(conn, false)
	if err != nil {
		return
	}

	contract, ok := obj.(*user.NodeContract)
	if !ok {
		return
	}

	// check user id
	if contract.UserID.IsZero() {
		return
	}

	// check node id
	if !contract.NodeID.IsEqual(mod.node.Identity()) {
		return
	}

	// don't sign contracts for less than an hour
	if !contract.ExpiresAt.Time().After(time.Now().Add(minimalContractLength)) {
		return
	}

	// sign the contract
	signed := &user.SignedNodeContract{
		NodeContract: contract,
	}

	signed.NodeSig, err = mod.Keys.SignASN1(mod.node.Identity(), signed.Hash())
	if err != nil {
		return
	}

	// write the node signature
	_, err = signed.NodeSig.WriteTo(conn)
	if err != nil {
		return
	}

	// read the user signature
	_, err = signed.UserSig.ReadFrom(conn)
	if err != nil {
		return
	}

	err = mod.SaveSignedNodeContract(signed)
	if err != nil {
		return
	}

	return mod.SetActiveContract(signed)
}
