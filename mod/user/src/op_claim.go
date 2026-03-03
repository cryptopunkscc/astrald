package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opClaimArgs struct {
	Target string
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpClaim(ctx *astral.Context, q *ops.Query, args opClaimArgs) (err error) {
	// get the active contract
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	if !q.Caller().IsEqual(ac.UserID) {
		return q.RejectWithCode(3)
	}

	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	nodeID, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// invite the node to sign a contract
	signed, err := mod.InviteNode(ctx, nodeID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// store the signed contract
	signedID, err := mod.Objects.Store(mod.ctx, mod.Objects.WriteDefault(), signed)
	if err != nil {
		return
	}

	mod.log.Info("signed contract %v with %v", signedID, nodeID)

	_, err = mod.IndexSignedNodeContract(signed)
	if err != nil {
		mod.log.Logv(1, "error indexing signed contract: %v", err)
	}

	return ch.Send(signed)
}
