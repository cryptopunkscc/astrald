package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	natclient "github.com/cryptopunkscc/astrald/mod/nat/client"
)

type opStartTraversal struct {
	Target string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpStartTraversal(ctx *astral.Context, q *ops.Query, args opStartTraversal) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	localIP, err := mod.getLocalIPv4()
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if args.Target != "" {
		// Initiator flow
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		mod.log.Log("starting traversal as initiator to %v", target)
		client := natclient.New(target, astrald.Default())
		punch, err := mod.openPuncher(nil)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		defer punch.Close()

		traversalClient, err := client.NewTraversalClient(localIP, punch.Session())
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		pair, err := traversalClient.StartTraversal(ctx, target)
		if err != nil {
			mod.log.Error("NAT traversal failed with %v: %v", target, err)
			return ch.Send(astral.Err(err))
		}

		mod.log.Info("NAT traversal § with %v: %v <-> %v", target, pair.PeerA.Endpoint, pair.PeerB.Endpoint)
		mod.addTraversedPair(*pair, true)
		return ch.Send(pair)
	}

	// Participant flow
	mod.log.Log("starting traversal as participant with %v", q.Caller())
	// mod.addTraversedPair(pair, false)
	return nil
}
