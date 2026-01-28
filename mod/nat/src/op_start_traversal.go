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
	exchange := nat.NewPunchExchange(ch)
	offer, err := exchange.Expect(nat.PunchSignalTypeOffer)
	if err != nil {
		return err
	}

	puncher, err := mod.openPuncher(offer.Session)
	if err != nil {
		return err
	}

	defer puncher.Close()

	err = exchange.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeAnswer,
		Session: puncher.Session(),
		IP:      localIP,
		Port:    astral.Uint16(puncher.LocalPort()),
	})
	if err != nil {
		return err
	}

	_, err = exchange.Expect(nat.PunchSignalTypeReady)
	if err != nil {
		return err
	}

	err = exchange.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeGo,
		Session: puncher.Session(),
	})
	if err != nil {
		return err
	}

	punchResult, err := puncher.HolePunch(ctx, offer.IP, int(offer.Port))
	if err != nil {
		return err
	}

	pair := nat.TraversedPortPair{
		PeerA: nat.PeerEndpoint{Identity: ctx.Identity()},
		PeerB: nat.PeerEndpoint{Identity: q.Caller(), Endpoint: nat.UDPEndpoint{IP: punchResult.RemoteIP, Port: punchResult.RemotePort}},
	}

	err = exchange.Send(&nat.PunchSignal{
		Signal:    nat.PunchSignalTypeResult,
		Session:   puncher.Session(),
		IP:        pair.PeerB.Endpoint.IP,
		Port:      pair.PeerB.Endpoint.Port,
		PairNonce: pair.Nonce,
	})
	if err != nil {
		return err
	}

	result, err := exchange.Expect(nat.PunchSignalTypeResult)
	if err != nil {
		return err
	}

	pair.Nonce = result.PairNonce
	pair.PeerA.Endpoint = nat.UDPEndpoint{IP: result.IP, Port: result.Port}

	mod.log.Info("NAT traversal § with %v: %v <-> %v", q.Caller(), pair.PeerA.Endpoint, pair.PeerB.Endpoint)
	mod.addTraversedPair(pair, false)
	return nil
}
