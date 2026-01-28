package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nat"
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
		punch, err := mod.openPuncher()
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		if _, err := punch.Open(); err != nil {
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

	puncher, err := mod.openPuncher()
	if err != nil {
		return err
	}

	traversal := nat.NewTraversal(
		ch,
		ctx.Identity(),
		q.Caller(),
		localIP,
		puncher,
	)

	err = ch.Switch(
		traversal.ExpectPunchSignal(nat.PunchSignalTypeOffer, traversal.OnOffer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if traversal.Puncher == nil {
		return fmt.Errorf("missing puncher after offer")
	}
	defer traversal.Puncher.Close()

	// 2. Ready → Go → Punch → Result
	err = ch.Switch(
		traversal.ExpectPunchSignal(nat.PunchSignalTypeReady, traversal.OnReady(ctx)),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	// 3. Final Result
	err = ch.Switch(
		traversal.ExpectPunchSignal(nat.PunchSignalTypeResult, traversal.OnResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	mod.log.Info(
		"NAT traversal § with %v: %v <-> %v",
		q.Caller(),
		traversal.Pair.PeerA.Endpoint,
		traversal.Pair.PeerB.Endpoint,
	)

	mod.addTraversedPair(traversal.Pair, false)
	return nil
}
