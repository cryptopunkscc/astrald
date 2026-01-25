package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/ip"
	natclient "github.com/cryptopunkscc/astrald/mod/nat/client"
)

type opStartTraversal struct {
	Target string `query:"optional"` // if not empty, act as initiator
	Out    string `query:"optional"`
}

func (mod *Module) OpStartTraversal(ctx *astral.Context, q *ops.Query, args opStartTraversal) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	ips := mod.IP.PublicIPCandidates()
	if len(ips) == 0 {
		return ch.Send(astral.NewError("no suitable IP addresses found"))
	}

	// Filter out IPv6 addresses, keep only IPv4
	var ipv4s []ip.IP
	for _, addr := range ips {
		if addr.IsIPv4() {
			ipv4s = append(ipv4s, addr)
		}
	}

	ips = ipv4s
	if len(ips) == 0 {
		return ch.Send(astral.NewError("no suitable IPv4 addresses found"))
	}

	if args.Target != "" {
		// TraversalRoleInitiator logic
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		mod.log.Log("starting traversal as initiator to %v", target)
		client := natclient.New(target, astrald.Default())
		peerCh, err := client.StartTraversalCh(ctx, args.Out)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		defer peerCh.Close()

		pair, err := mod.Traverse(ctx, peerCh, TraversalRoleInitiator, target, ips[0])
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		mod.addTraversedPair(pair, true)

		if err = ch.Send(&pair); err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		return nil
	}

	mod.log.Log("starting traversal as responder with %v", q.Caller())

	pair, err := mod.Traverse(ctx, ch, TraversalRoleParticipant, q.Caller(), ips[0])
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	mod.addTraversedPair(pair, false)
	return nil
}
