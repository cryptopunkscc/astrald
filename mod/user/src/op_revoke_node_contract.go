package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opRevokeNodeContractArgs struct {
	ContractId *astral.ObjectID

	StartsAt *astral.Time    `query:"optional"` // absolute timestamp
	StartsIn astral.Duration `query:"optional"` // relative duration

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRevokeNodeContract(ctx *astral.Context, q shell.Query, args opRevokeNodeContractArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	var startsAt astral.Time
	switch {
	case args.StartsAt != nil:
		startsAt = *args.StartsAt
	case args.StartsIn != 0:
		startsAt = astral.Time(astral.Now().Time().Add(args.StartsIn.Duration()))
	default:
		startsAt = astral.Now() // default immediate start
	}

	nodeContract, err := mod.GetNodeContract(args.ContractId)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	if startsAt.Time().After(nodeContract.ExpiresAt.Time()) {
		return ch.Write(astral.NewError(user.ErrNodeContractRevocationForExpiredContract.Error()))
	}

	if nodeContract.IsExpired() {
		return ch.Write(astral.NewError(user.ErrNodeContractAlreadyExpired.Error()))
	}

	var revocation = &user.NodeContractRevocation{
		ContractID: args.ContractId,
		StartsAt:   startsAt,
		ExpiresAt:  nodeContract.ExpiresAt,
	}

	var signed = &user.SignedNodeContractRevocation{
		NodeContractRevocation: revocation,
	}

	// contract revocation must be signed by user key
	signed.UserSig, err = mod.Keys.SignASN1(nodeContract.UserID, signed.Hash())
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	err = mod.SaveSignedRevocationContract(signed)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return
}
