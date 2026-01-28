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

	ExpectPunchSignal := func(signalType astral.String8, on func(sig *nat.PunchSignal) (err error)) func(sig *nat.PunchSignal) (err error) {
		return func(sig *nat.PunchSignal) (err error) {
			if sig.Signal != signalType {
				return fmt.Errorf("unexpected punch signal: %v", sig.Signal)
			}

			return on(sig)
		}
	}

	// Participant flow
	mod.log.Log("starting traversal as participant with %v", q.Caller())

	tr := nat.NewTraversal(ch)

	var (
		offer   *nat.PunchSignal
		puncher nat.Puncher
		pair    nat.TraversedPortPair
	)

	// Participant flow
	mod.log.Log("starting traversal as participant with %v", q.Caller())

	onOffer := func(sig *nat.PunchSignal) error {
		offer = sig

		var err error
		puncher, err = mod.openPuncher(sig.Session)
		if err != nil {
			return err
		}

		return tr.SendAnswer(
			localIP,
			uint16(puncher.LocalPort()),
			puncher.Session(),
		)
	}

	onReady := func(_ *nat.PunchSignal) error {
		if err := tr.SendGo(); err != nil {
			return err
		}

		res, err := puncher.HolePunch(ctx, offer.IP, int(offer.Port))
		if err != nil {
			return err
		}

		pair = nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: ctx.Identity(),
			},
			PeerB: nat.PeerEndpoint{
				Identity: q.Caller(),
				Endpoint: nat.UDPEndpoint{
					IP:   res.RemoteIP,
					Port: res.RemotePort,
				},
			},
		}

		return tr.SendResult(
			pair.PeerB.Endpoint.IP,
			uint16(pair.PeerB.Endpoint.Port),
			pair.Nonce,
		)
	}

	onResult := func(sig *nat.PunchSignal) error {
		pair.Nonce = sig.PairNonce
		pair.PeerA.Endpoint = nat.UDPEndpoint{
			IP:   sig.IP,
			Port: sig.Port,
		}

		mod.log.Info(
			"NAT traversal § with %v: %v <-> %v",
			q.Caller(),
			pair.PeerA.Endpoint,
			pair.PeerB.Endpoint,
		)

		mod.addTraversedPair(pair, false)
		return nil
	}

	err = ch.Switch(
		ExpectPunchSignal(nat.PunchSignalTypeOffer, onOffer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	defer puncher.Close()

	// 2. Ready → Go → Punch → Result
	err = ch.Switch(
		ExpectPunchSignal(nat.PunchSignalTypeReady, onReady),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	// 3. Final Result
	err = ch.Switch(
		ExpectPunchSignal(nat.PunchSignalTypeResult, onResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	return nil
}
