package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStartTraversal struct {
	Target string `query:"optional"` // if not empty, act as initiator
	Out    string `query:"optional"`
}

func (mod *Module) OpStartTraversal(ctx *astral.Context, q shell.Query, args opStartTraversal) error {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	ips := mod.IP.PublicIPCandidates()
	if len(ips) == 0 {
		return ch.Write(astral.NewError("no suitable IP addresses found"))
	}

	if args.Target != "" {
		// TraversalRoleInitiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		mod.log.Log("starting traversal as initiator to %v", target)
		peerCh, err := query.RouteChan(ctx.IncludeZone(astral.ZoneNetwork), mod.node, query.New(ctx.Identity(), target, nat.MethodStartNatTraversal, &opStartTraversal{
			Out: args.Out,
		}))
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer peerCh.Close()

		pair, err := mod.Traverse(ctx, peerCh, TraversalRoleInitiator, target, ips[0])
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		mod.addTraversedPair(pair, true)

		if err = ch.Write(&pair); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	mod.log.Log("starting traversal as responder with %v", q.Caller())

	pair, err := mod.Traverse(ctx, ch, TraversalRoleParticipant, q.Caller(), ips[0])
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	mod.addTraversedPair(pair, false)
	return nil
}
