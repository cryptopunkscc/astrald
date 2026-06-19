package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opAdoptArgs struct {
	Target string
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

// OpAdopt adopts a target node into the active contract and indexes the signed result.
// Requires an active contract; caller must be the contract issuer (code 3 otherwise).
// Pushes the signed contract to the local swarm asynchronously after indexing.
func (mod *Module) OpAdopt(ctx *astral.Context, q *routing.IncomingQuery, args opAdoptArgs) (err error) {
	// get the active contract
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	if !q.Caller().IsEqual(ac.Issuer) {
		return q.RejectWithCode(3)
	}

	ch := q.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	nodeID, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// issue a membership contract for the node
	signed, err := mod.IssueMembership(ctx, nodeID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.Auth.IndexContract(ctx, signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	_, err = mod.Objects.Store(ctx, mod.Objects.WriteDefault(), signed)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	go mod.PushToLocalSwarm(mod.ctx, signed)

	// why: PushToLocalSwarm only sends the new contract, leaving the invitee
	// without the inviter's own and sibling contracts. The LinkCreatedEvent
	// trigger already fired before indexing, so sync the joined node here.
	mod.Scheduler.Schedule(mod.NewSyncNodesTask(signed.Subject))

	mod.log.Info("signed contract with %v", nodeID)
	return ch.Send(signed)
}
