package nat

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type traversal struct {
	log           *log.Logger
	localIdentity *astral.Identity // localIdentity helps us select right endpoints
	localPublicIP ip.IP
	role          TraversalRole
	ch            *channel.Channel
	peer          *astral.Identity
	//
	punch nat.Puncher
	pair  nat.TraversedPortPair
}

func (mod *Module) Traverse(
	ctx context.Context,
	ch *channel.Channel,
	role TraversalRole,
	target *astral.Identity,
	publicIP ip.IP,
) (nat.TraversedPortPair, error) {
	t := &traversal{
		log:           mod.log,
		localIdentity: mod.ctx.Identity(),
		role:          role,
		ch:            ch,
		peer:          target,
		localPublicIP: publicIP,
	}

	pair, err := t.run(ctx)
	if err != nil {
		mod.log.Error("NAT traversal failed with %v: %v", target, err)
		return nat.TraversedPortPair{}, err
	}

	mod.log.Info("NAT traversal succeeded with %v: %v <-> %v", target, pair.PeerA.Endpoint, pair.PeerB.Endpoint)
	return pair, err
}

type TraversalRole int

const (
	TraversalRoleInitiator TraversalRole = iota
	TraversalRoleParticipant
)

func (t *traversal) run(ctx context.Context) (nat.TraversedPortPair, error) {
	defer t.close()

	switch t.role {
	case TraversalRoleInitiator:
		if err := t.initiatorFlow(ctx); err != nil {
			return nat.TraversedPortPair{}, err
		}
	case TraversalRoleParticipant:
		if err := t.participantFlow(ctx); err != nil {
			return nat.TraversedPortPair{}, err
		}
	}

	t.pair.CreatedAt = astral.Time(time.Now())
	return t.pair, nil
}

func (t *traversal) initiatorFlow(ctx context.Context) error {
	if err := t.openPuncher(nil); err != nil {
		return err
	}

	if err := t.ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeOffer,
		Session: t.punch.Session(),
		IP:      t.localPublicIP,
		Port:    astral.Uint16(t.punch.LocalPort()),
	}); err != nil {
		return err
	}

	ans, err := t.receiveSignal(nat.PunchSignalTypeAnswer)
	if err != nil {
		return err
	}

	if err := t.ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeReady,
		Session: t.punch.Session(),
	}); err != nil {
		return err
	}

	if _, err := t.receiveSignal(nat.PunchSignalTypeGo); err != nil {
		return err
	}

	res, err := t.punch.HolePunch(ctx, ans.IP, int(ans.Port))
	if err != nil {
		return err
	}
	t.pair = t.makePair(res.RemoteIP, res.RemotePort)

	t.pair.Nonce = astral.NewNonce()
	if err := t.ch.Send(&nat.PunchSignal{
		Signal:    nat.PunchSignalTypeResult,
		Session:   t.punch.Session(),
		IP:        t.pair.PeerB.Endpoint.IP,
		Port:      t.pair.PeerB.Endpoint.Port,
		PairNonce: t.pair.Nonce,
	}); err != nil {
		return err
	}

	resSig, err := t.receiveSignal(nat.PunchSignalTypeResult)
	if err != nil {
		return err
	}
	t.pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	return nil
}

func (t *traversal) participantFlow(ctx context.Context) error {
	offer, err := t.receiveSignal(nat.PunchSignalTypeOffer)
	if err != nil {
		return err
	}

	if err := t.openPuncher(offer.Session); err != nil {
		return err
	}

	if err := t.ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeAnswer,
		Session: t.punch.Session(),
		IP:      t.localPublicIP,
		Port:    astral.Uint16(t.punch.LocalPort()),
	}); err != nil {
		return err
	}

	if _, err := t.receiveSignal(nat.PunchSignalTypeReady); err != nil {
		return err
	}

	if err := t.ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeGo,
		Session: t.punch.Session(),
	}); err != nil {
		return err
	}

	res, err := t.punch.HolePunch(ctx, offer.IP, int(offer.Port))
	if err != nil {
		return err
	}
	t.pair = t.makePair(res.RemoteIP, res.RemotePort)

	resSig, err := t.receiveSignal(nat.PunchSignalTypeResult)
	if err != nil {
		return err
	}

	t.pair.Nonce = resSig.PairNonce
	t.pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	return t.ch.Send(&nat.PunchSignal{
		Signal:    nat.PunchSignalTypeResult,
		Session:   t.punch.Session(),
		IP:        t.pair.PeerB.Endpoint.IP,
		Port:      t.pair.PeerB.Endpoint.Port,
		PairNonce: t.pair.Nonce,
	})
}

func (t *traversal) openPuncher(session []byte) error {
	cb := &ConePuncherCallbacks{
		OnAttempt:       func(peer ip.IP, port int, _ []*net.UDPAddr) { t.log.Log("punching → %v:%v", peer, port) },
		OnProbeReceived: func(from *net.UDPAddr) { t.log.Log("probe ← %v", from) },
	}
	p, err := newConePuncher(session, cb)
	if err != nil {
		return err
	}
	if _, err = p.Open(); err != nil {
		return err
	}
	t.punch = p
	return nil
}

func (t *traversal) makePair(ip ip.IP, port astral.Uint16) nat.TraversedPortPair {
	return nat.TraversedPortPair{
		PeerA: nat.PeerEndpoint{Identity: t.localIdentity},
		PeerB: nat.PeerEndpoint{Identity: t.peer, Endpoint: nat.UDPEndpoint{IP: ip, Port: port}},
	}
}

func (t *traversal) receiveSignal(expected astral.String8) (*nat.PunchSignal, error) {
	obj, err := t.ch.Receive()
	if err != nil {
		return nil, err
	}

	switch sig := obj.(type) {
	case *astral.ErrorMessage:
		return nil, fmt.Errorf("received error message: %s", sig.Error())
	case *nat.PunchSignal:
		if sig.Signal != expected {
			return nil, fmt.Errorf("expected %s, got %s", expected, sig.Signal)
		}

		if t.punch != nil {
			if !bytes.Equal(sig.Session, t.punch.Session()) {
				return nil, fmt.Errorf("session mismatch")
			}
		}

		return sig, nil
	default:
		return nil, fmt.Errorf("unexpected message type: %T", sig)
	}
}

func (t *traversal) close() {
	if t.punch != nil {
		_ = t.punch.Close()
	}
}
