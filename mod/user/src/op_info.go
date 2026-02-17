package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opInfoArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpInfo(ctx *astral.Context, q *ops.Query, args opInfoArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	var foundInSwarm bool
	for _, swarm := range mod.LocalSwarm() {
		if swarm.IsEqual(q.Caller()) {
			foundInSwarm = true
		}
	}

	if !foundInSwarm && !q.Caller().IsEqual(ac.UserID) {
		return q.Reject()
	}

	ch := q.AcceptChannel(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	contractID, _ := astral.ResolveObjectID(ac)

	return ch.Send(&user.Info{
		NodeAlias:  astral.String8(mod.Dir.DisplayName(ac.NodeID)),
		UserAlias:  astral.String8(mod.Dir.DisplayName(ac.UserID)),
		ContractID: contractID,
		Contract:   ac,
	})
}
