package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opExpelArgs struct {
	Target string
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

// OpExpel permanently bans the target node from the swarm and returns the signed ban.
// Requires an active contract; caller must be the contract issuer (code 3 otherwise).
func (mod *Module) OpExpel(ctx *astral.Context, q *routing.IncomingQuery, args opExpelArgs) (err error) {
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

	signed, err := mod.Expel(ctx, nodeID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(signed)
}
