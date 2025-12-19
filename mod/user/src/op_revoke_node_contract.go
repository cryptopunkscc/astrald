package user

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opRevokeNodeContractArgs struct {
	ContractId *astral.ObjectID
	RevokeAs   string `query:"optional"` // default points to user
	In         string `query:"optional"`
	Out        string `query:"optional"`
}

func (mod *Module) OpRevokeNodeContract(ctx *astral.Context, q shell.Query, args opRevokeNodeContractArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		// cannot handle if we dont have active contract
		return q.RejectWithCode(2)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	if args.RevokeAs == "" {
		args.RevokeAs = "user"
	}

	nodeContract, err := mod.GetNodeContract(args.ContractId)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	if !nodeContract.UserID.IsEqual(ac.UserID) {
		return ch.Write(user.ErrNodeContractRevocationInvalid)
	}

	if nodeContract.IsExpired() {
		return ch.Write(user.ErrNodeContractAlreadyExpired)
	}

	var revocation = &user.NodeContractRevocation{
		ContractID: args.ContractId,
		ExpiresAt:  astral.Time(nodeContract.ExpiresAt.Time().Add(minimalRevocationLength)),
		CreatedAt:  astral.Time(time.Now()),
	}

	var signed = &user.SignedNodeContractRevocation{
		NodeContractRevocation: revocation,
	}

	var revoker *user.Revoker
	switch args.RevokeAs {
	case "user":
		userSig, err := mod.Keys.SignASN1(nodeContract.UserID, signed.Hash())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		revoker = &user.Revoker{
			ID:  ac.UserID,
			Sig: userSig,
		}
	case "node":
		nodeSig, err := mod.Keys.SignASN1(mod.ctx.Identity(), signed.Hash())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		revoker = &user.Revoker{
			ID:  nodeContract.NodeID,
			Sig: nodeSig,
		}

		err = signed.Attachments.Append(mod.ActiveContract())
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	default:
		return ch.Write(astral.NewError(fmt.Errorf(`invalid revoke-as "%s"`, args.RevokeAs).Error()))
	}

	signed.Revoker = revoker

	err = mod.SaveSignedRevocationContract(signed, nodeContract)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return
}
