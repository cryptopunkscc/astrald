package user

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type opClaimArgs struct {
	Target string
}

// OpClaim invites the target node to sign a contract with the current user
func (mod *Module) OpClaim(ctx *astral.Context, q shell.Query, args opClaimArgs) (err error) {
	// we need an active contract to claim other nodes
	ac := mod.ActiveContract()
	if ac == nil {
		return q.Reject()
	}

	// only the current user can perform this op
	if !q.Caller().IsEqual(ac.UserID) {
		return q.Reject()
	}

	conn := q.Accept()
	defer conn.Close()
	enc := json.NewEncoder(conn)
	enc.SetIndent("", "  ")

	// resolve target identity
	targetID, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// send an invite to the target node
	invite, err := query.Run(mod.node, targetID, "user.invite", nil)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	signed := &user.SignedNodeContract{
		NodeContract: &user.NodeContract{
			UserID:    ac.UserID,
			NodeID:    targetID,
			ExpiresAt: astral.Time(time.Now().Add(defaultContractValidity)),
		},
	}

	// write a proposed *NodeContract
	_, err = astral.Write(invite, signed.NodeContract, false)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// read node signature
	_, err = signed.NodeSig.ReadFrom(invite)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// verify node signature
	err = mod.Keys.VerifyASN1(signed.NodeID, signed.Hash(), signed.NodeSig)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// generate user signature
	signed.UserSig, err = mod.Keys.SignASN1(signed.UserID, signed.Hash())
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// write user signature
	_, err = signed.UserSig.WriteTo(invite)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// save the fully signed contract
	err = mod.SaveSignedNodeContract(signed)
	if err != nil {
		return enc.Encode(map[string]interface{}{
			"error": err.Error(),
		})
	}

	return enc.Encode(map[string]interface{}{
		"status":   "ok",
		"contract": signed,
	})
}
