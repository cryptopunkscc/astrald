package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	modnat "github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStartTraversal struct {
	// Active side fields
	Target string `query:"optional"` // if not empty act as initiator
	Out    string `query:"optional"`
}

func (mod *Module) OpStartTraversal(ctx *astral.Context, q shell.Query, args opStartTraversal) error {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer func() { _ = ch.Close() }()

	ips := mod.IP.PublicIPCandidates()
	if len(ips) == 0 {
		return ch.Write(astral.NewError("no suitable IP addresses found"))
	}

	if args.Target != "" {
		// Initiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		peerCh, err := query.RouteChan(ctx.IncludeZone(astral.ZoneNetwork), mod.node, query.New(ctx.Identity(), target, modnat.MethodStartNatTraversal, &opStartTraversal{
			Out: args.Out,
		}))
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer func() { _ = peerCh.Close() }()

		// run minimal state machine as initiator over peerCh
		var sm = traversal{role: RoleInitiator, ch: peerCh, ips: ips, selfID: ctx.Identity(), peerID: target}
		pair, err := sm.Run(ctx)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		mod.addTraversedPair(pair)

		if err = ch.Write(&pair); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return nil
	}

	// Responder logic via state machine on ch
	var sm = traversal{role: RoleResponder, ch: ch, ips: ips, selfID: ctx.Identity(), peerID: q.Caller()}
	pair, err := sm.Run(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	mod.addTraversedPair(pair)

	return nil
}
