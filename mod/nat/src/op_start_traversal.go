package nat

import (
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

		puncher, err := mod.openPuncher()
		if err != nil {
			return ch.Send(astral.Err(err))
		}
		defer puncher.Close()

		client := natclient.New(target, astrald.Default())
		pair, err := client.StartTraversal(ctx, target, localIP, puncher)
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
	defer puncher.Close()

	traversal := nat.NewTraversal(ctx.Identity(), q.Caller(), localIP)

	// 1. Receive offer
	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeOffer, traversal.OnOffer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err := puncher.SetSession(traversal.Session); err != nil {
		return err
	}

	localPort, err := puncher.Open()
	if err != nil {
		return err
	}
	traversal.LocalPort = astral.Uint16(localPort)

	if err := ch.Send(traversal.AnswerSignal()); err != nil {
		return err
	}

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeReady, nil),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err := ch.Send(traversal.GoSignal()); err != nil {
		return err
	}

	result, err := puncher.HolePunch(ctx, traversal.PeerIP, int(traversal.PeerPort))
	if err != nil {
		return err
	}

	traversal.SetPunchResult(result)
	if err := ch.Send(traversal.ResultSignal()); err != nil {
		return err
	}
	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeResult, traversal.OnResult),
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
