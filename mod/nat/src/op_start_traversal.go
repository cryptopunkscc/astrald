package nat

import (
	"context"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/ip"
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
		peerCh, err := client.StartTraversalCh(ctx, "")
		if err != nil {
			return ch.Send(astral.Err(err))
		}
		defer peerCh.Close()

		pair, err := mod.initiatorFlow(ctx, peerCh, target, localIP)
		if err != nil {
			mod.log.Error("NAT traversal failed with %v: %v", target, err)
			return ch.Send(astral.Err(err))
		}

		mod.log.Info("NAT traversal succeeded with %v: %v <-> %v", target, pair.PeerA.Endpoint, pair.PeerB.Endpoint)
		mod.addTraversedPair(pair, true)
		return ch.Send(&pair)
	}

	// Participant flow
	mod.log.Log("starting traversal as participant with %v", q.Caller())

	pair, err := mod.participantFlow(ctx, ch, q.Caller(), localIP)
	if err != nil {
		mod.log.Error("NAT traversal failed with %v: %v", q.Caller(), err)
		return ch.Send(astral.Err(err))
	}

	mod.log.Info("NAT traversal succeeded with %v: %v <-> %v", q.Caller(), pair.PeerA.Endpoint, pair.PeerB.Endpoint)
	mod.addTraversedPair(pair, false)
	return nil
}

func (mod *Module) initiatorFlow(ctx context.Context, ch *channel.Channel, peer *astral.Identity, localIP ip.IP) (nat.TraversedPortPair, error) {
	punch, err := mod.openPuncher(nil)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}
	defer punch.Close()

	exchange := nat.NewPunchExchange(ch)
	exchange.Session = punch.Session()

	// Offer
	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeOffer, IP: localIP, Port: astral.Uint16(punch.LocalPort())}); err != nil {
		return nat.TraversedPortPair{}, err
	}
	ans, err := exchange.Expect(nat.PunchSignalTypeAnswer)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}

	// Ready/Go
	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeReady}); err != nil {
		return nat.TraversedPortPair{}, err
	}
	if _, err := exchange.Expect(nat.PunchSignalTypeGo); err != nil {
		return nat.TraversedPortPair{}, err
	}

	// Punch
	res, err := punch.HolePunch(ctx, ans.IP, int(ans.Port))
	if err != nil {
		return nat.TraversedPortPair{}, err
	}

	pair := nat.TraversedPortPair{
		Nonce:     astral.NewNonce(),
		CreatedAt: astral.Time(time.Now()),
		PeerA:     nat.PeerEndpoint{Identity: mod.ctx.Identity()},
		PeerB:     nat.PeerEndpoint{Identity: peer, Endpoint: nat.UDPEndpoint{IP: res.RemoteIP, Port: res.RemotePort}},
	}

	// Exchange results
	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeResult, IP: pair.PeerB.Endpoint.IP, Port: pair.PeerB.Endpoint.Port, PairNonce: pair.Nonce}); err != nil {
		return nat.TraversedPortPair{}, err
	}
	resSig, err := exchange.Expect(nat.PunchSignalTypeResult)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}
	pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	return pair, nil
}

func (mod *Module) participantFlow(ctx context.Context, ch *channel.Channel, peer *astral.Identity, localIP ip.IP) (nat.TraversedPortPair, error) {
	exchange := nat.NewPunchExchange(ch)

	// Offer
	offer, err := exchange.Expect(nat.PunchSignalTypeOffer)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}

	punch, err := mod.openPuncher(offer.Session)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}
	defer punch.Close()

	exchange.Session = punch.Session()

	// Answer
	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeAnswer, IP: localIP, Port: astral.Uint16(punch.LocalPort())}); err != nil {
		return nat.TraversedPortPair{}, err
	}

	// Ready/Go
	if _, err := exchange.Expect(nat.PunchSignalTypeReady); err != nil {
		return nat.TraversedPortPair{}, err
	}
	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeGo}); err != nil {
		return nat.TraversedPortPair{}, err
	}

	// Punch
	res, err := punch.HolePunch(ctx, offer.IP, int(offer.Port))
	if err != nil {
		return nat.TraversedPortPair{}, err
	}

	pair := nat.TraversedPortPair{
		CreatedAt: astral.Time(time.Now()),
		PeerA:     nat.PeerEndpoint{Identity: mod.ctx.Identity()},
		PeerB:     nat.PeerEndpoint{Identity: peer, Endpoint: nat.UDPEndpoint{IP: res.RemoteIP, Port: res.RemotePort}},
	}

	// Exchange results
	resSig, err := exchange.Expect(nat.PunchSignalTypeResult)
	if err != nil {
		return nat.TraversedPortPair{}, err
	}
	pair.Nonce = resSig.PairNonce
	pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	if err := exchange.Send(&nat.PunchSignal{Signal: nat.PunchSignalTypeResult, IP: pair.PeerB.Endpoint.IP, Port: pair.PeerB.Endpoint.Port, PairNonce: pair.Nonce}); err != nil {
		return nat.TraversedPortPair{}, err
	}

	return pair, nil
}

func (mod *Module) getLocalIPv4() (ip.IP, error) {
	for _, addr := range mod.IP.PublicIPCandidates() {
		if addr.IsIPv4() {
			return addr, nil
		}
	}
	return nil, nat.ErrNoSuitableIP
}
