package nat

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type traversal struct {
	log   *log.Logger
	role  TraversalRole
	ch    *astral.Channel
	local *astral.Identity
	peer  *astral.Identity
	pubIP ip.IP
	punch nat.Puncher
	pair  nat.TraversedPortPair
}

type TraversalRole bool

const (
	Initiator TraversalRole = true
	Responder TraversalRole = false
)

func Traverse(ctx context.Context, ch *astral.Channel, local, peer *astral.Identity, role TraversalRole, pubIP ip.IP, log *log.Logger) (nat.TraversedPortPair, error) {
	t := &traversal{
		log:   log,
		role:  role,
		ch:    ch,
		local: local,
		peer:  peer,
		pubIP: pubIP,
	}

	pair, err := t.run(ctx)
	if err != nil {
		log.Error("NAT traversal failed with %v: %v", peer, err)
	} else {
		log.Info("direct pair %v", &pair)
	}
	return pair, err
}

func (t *traversal) run(ctx context.Context) (nat.TraversedPortPair, error) {
	defer t.close()

	if Initiator == t.role {
		if err := t.initiatorFlow(ctx); err != nil {
			return nat.TraversedPortPair{}, err
		}
	} else {
		if err := t.responderFlow(ctx); err != nil {
			return nat.TraversedPortPair{}, err
		}
	}

	t.pair.CreatedAt = astral.Time(time.Now())
	return t.pair, nil
}

// ——————————————————————— Initiator ———————————————————————
func (t *traversal) initiatorFlow(ctx context.Context) error {
	if err := t.openPuncher(nil); err != nil {
		return err
	}

	if err := t.send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeOffer,
		Session: t.punch.Session(),
		IP:      t.pubIP,
		Port:    astral.Uint16(t.punch.LocalPort()),
	}); err != nil {
		return err
	}

	ans, err := t.recvType(nat.PunchSignalTypeAnswer)
	if err != nil {
		return err
	}

	if err := t.send(&nat.PunchSignal{Signal: nat.PunchSignalTypeReady}); err != nil {
		return err
	}
	if _, err := t.recvType(nat.PunchSignalTypeGo); err != nil {
		return err
	}

	res, err := t.punch.HolePunch(ctx, ans.IP, int(ans.Port))
	if err != nil {
		return err
	}
	t.pair = t.makePair(res.RemoteIP, res.RemotePort)

	t.pair.Nonce = astral.NewNonce()
	if err := t.send(&nat.PunchSignal{
		Signal:    nat.PunchSignalTypeResult,
		IP:        t.pair.PeerB.Endpoint.IP,
		Port:      t.pair.PeerB.Endpoint.Port,
		PairNonce: t.pair.Nonce,
	}); err != nil {
		return err
	}

	resSig, err := t.recvType(nat.PunchSignalTypeResult)
	if err != nil {
		return err
	}
	t.pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	return nil
}

// ——————————————————————— Responder ———————————————————————
func (t *traversal) responderFlow(ctx context.Context) error {
	offer, err := t.recvType(nat.PunchSignalTypeOffer)
	if err != nil {
		return err
	}

	if err := t.openPuncher(offer.Session); err != nil {
		return err
	}

	if err := t.send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeAnswer,
		Session: t.punch.Session(),
		IP:      t.pubIP,
		Port:    astral.Uint16(t.punch.LocalPort()),
	}); err != nil {
		return err
	}

	if _, err := t.recvType(nat.PunchSignalTypeReady); err != nil {
		return err
	}
	if err := t.send(&nat.PunchSignal{Signal: nat.PunchSignalTypeGo}); err != nil {
		return err
	}

	res, err := t.punch.HolePunch(ctx, offer.IP, int(offer.Port))
	if err != nil {
		return err
	}
	t.pair = t.makePair(res.RemoteIP, res.RemotePort)

	resSig, err := t.recvType(nat.PunchSignalTypeResult)
	if err != nil {
		return err
	}

	t.pair.Nonce = resSig.PairNonce
	t.pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}

	return t.send(&nat.PunchSignal{
		Signal:    nat.PunchSignalTypeResult,
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
		PeerA: nat.PeerEndpoint{Identity: t.local},
		PeerB: nat.PeerEndpoint{Identity: t.peer, Endpoint: nat.UDPEndpoint{IP: ip, Port: port}},
	}
}

func (t *traversal) send(sig *nat.PunchSignal) error {
	sig.Session = t.punch.Session()
	return t.ch.Write(sig)
}

func (t *traversal) recv() (*nat.PunchSignal, error) {
	obj, err := t.ch.Read()
	if err != nil {
		return nil, err
	}
	sig, ok := obj.(*nat.PunchSignal)
	if !ok {
		return nil, fmt.Errorf("expected PunchSignal, got %T", obj)
	}
	return sig, nil
}

func (t *traversal) recvType(expected astral.String8) (*nat.PunchSignal, error) {
	sig, err := t.recv()
	if err != nil {
		return nil, err
	}
	if sig.Signal != expected {
		return nil, fmt.Errorf("expected %s, got %s", expected, sig.Signal)
	}
	if !bytes.Equal(sig.Session, t.punch.Session()) {
		return nil, fmt.Errorf("session mismatch")
	}
	return sig, nil
}

func (t *traversal) close() {
	if t.punch != nil {
		_ = t.punch.Close()
	}
}
