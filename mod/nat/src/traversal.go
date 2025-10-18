package nat

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type Role int

const (
	RoleInitiator Role = iota
	RoleResponder
)

type State int

const (
	StateOfferExchange State = iota
	StateReadyPhase
	StatePunch
	StateResultExchange
	StateDone
)

type stateFn func(*astral.Context) (state State, err error)

// traversal is a tiny state machine driving NatSignal exchange.
type traversal struct {
	role Role
	ch   *astral.Channel // channel to exchange NatSignal

	localIdentity *astral.Identity
	peerIdentity  *astral.Identity
	localPublicIP ip.IP
	puncher       nat.Puncher
	// received signals
	offer      *nat.NatSignal
	answer     *nat.NatSignal
	goSignal   *nat.NatSignal
	peerResult *nat.NatSignal

	pair nat.EndpointPair
}

func (t *traversal) Run(ctx *astral.Context) (pair nat.EndpointPair, err error) {
	var handlers = map[State]stateFn{
		StateOfferExchange:  t.handleOfferExchange,
		StateReadyPhase:     t.handleReadyPhase,
		StatePunch:          t.handlePunch,
		StateResultExchange: t.handleResultExchange,
	}

	defer t.closePuncher()
	state := StateOfferExchange
	for state != StateDone {
		h, ok := handlers[state]
		if !ok {
			return pair, fmt.Errorf("invalid state")
		}

		next, hErr := h(ctx)
		if hErr != nil {
			err = hErr
			return pair, err
		}
		state = next
	}

	pair = t.pair
	pair.CreatedAt = astral.Time(time.Now())

	return pair, err
}

func (t *traversal) handleOfferExchange(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		// as an initiator, setup puncher and send offer
		err = t.setupPuncher(nil)
		if err != nil {
			return state, err
		}
		err = t.ch.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeOffer,
			Session: t.puncher.Session(),
			IP:      t.localPublicIP,
			Port:    astral.Uint16(t.puncher.LocalPort()),
		})
		if err != nil {
			return state, err
		}
		ans, err := t.recv()
		if err != nil {
			return state, err
		}
		if err = t.verify(ans, nat.NatSignalTypeAnswer); err != nil {
			return 0, err
		}
		t.answer = ans
		state = StateReadyPhase
		return state, err
	}

	off, err := t.recv()
	if err != nil {
		return state, err
	}
	if len(off.Session) == 0 {
		return 0, fmt.Errorf("missing session in offer")
	}
	t.offer = off
	err = t.setupPuncher(off.Session)
	if err != nil {
		return state, err
	}
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeAnswer,
		Session: t.puncher.Session(), IP: t.localPublicIP,
		Port: astral.Uint16(t.puncher.LocalPort())})
	if err != nil {
		return state, err
	}
	state = StateReadyPhase
	return state, err
}

func (t *traversal) handleReadyPhase(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		err = t.ch.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeReady,
			Session: t.puncher.Session(),
		})
		if err != nil {
			return state, err
		}
		goSig, err := t.recv()
		if err != nil {
			return state, err
		}
		if err = t.verify(goSig, nat.NatSignalTypeGo); err != nil {
			return 0, err
		}
		t.goSignal = goSig
		state = StatePunch
		return state, err
	}
	ready, err := t.recv()
	if err != nil {
		return state, err
	}
	if err = t.verify(ready, nat.NatSignalTypeReady); err != nil {
		return 0, err
	}
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeGo,
		Session: t.puncher.Session()})
	if err != nil {
		return state, err
	}
	state = StatePunch
	return state, err
}

func (t *traversal) handlePunch(ctx *astral.Context) (state State, err error) {
	var res *nat.PunchResult
	if t.role == RoleInitiator {
		res, err = t.puncher.HolePunch(ctx, t.answer.IP, int(t.answer.Port))
	} else {
		res, err = t.puncher.HolePunch(ctx, t.offer.IP, int(t.offer.Port))
	}
	if err != nil {
		return state, err
	}
	// assign the observed endpoint to the pair
	observedPeer := utp.Endpoint{IP: res.RemoteIP, Port: res.RemotePort}
	pair := nat.EndpointPair{
		PeerA: nat.PeerEndpoint{Identity: t.localIdentity},
		PeerB: nat.PeerEndpoint{Identity: t.peerIdentity, Endpoint: observedPeer},
	}
	t.pair = pair
	state = StateResultExchange
	return state, err
}

func (t *traversal) handleResultExchange(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		err = t.ch.Write(&nat.NatSignal{
			Signal:  nat.NatSignalTypeResult,
			Session: t.puncher.Session(),
			IP:      t.pair.PeerB.Endpoint.IP,
			Port:    t.pair.PeerB.Endpoint.Port,
		})
		if err != nil {
			return state, err
		}
		res, err := t.recv()
		if err != nil {
			return state, err
		}
		if err = t.verify(res, nat.NatSignalTypeResult); err != nil {
			return 0, err
		}
		t.pair.PeerA.Endpoint = utp.Endpoint{IP: res.IP, Port: res.Port}
		state = StateDone
		return state, err
	}

	res, err := t.recv()
	if err != nil {
		return state, err
	}

	err = t.verify(res, nat.NatSignalTypeResult)
	if err != nil {
		return 0, err
	}
	t.pair.PeerA.Endpoint = utp.Endpoint{IP: res.IP, Port: res.Port}
	// responder sends result back to initiator
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeResult, Session: t.puncher.Session(), IP: t.pair.PeerB.Endpoint.IP, Port: t.pair.PeerB.Endpoint.Port})
	if err != nil {
		return state, err
	}
	state = StateDone
	return state, err
}

func (t *traversal) setupPuncher(session []byte) error {
	var err error
	if len(session) > 0 {
		t.puncher, err = newConePuncherWithSession(session)
	} else {
		t.puncher, err = newConePuncher()
	}
	if err != nil {
		return err
	}

	_, err = t.puncher.Open()
	if err != nil {
		return err
	}

	return nil
}

func (t *traversal) closePuncher() {
	if t.puncher != nil {
		_ = t.puncher.Close()
	}
}

func (t *traversal) recv() (*nat.NatSignal, error) {
	obj, err := t.ch.Read()
	if err != nil {
		return nil, err
	}
	sig, ok := obj.(*nat.NatSignal)
	if !ok {
		return nil, fmt.Errorf("unexpected object type")
	}
	return sig, nil
}

func (t *traversal) verify(sig *nat.NatSignal, expected string) error {
	if sig == nil || sig.Signal != astral.String8(expected) {
		return fmt.Errorf("invalid %s signal", expected)
	}
	if !bytes.Equal(sig.Session, t.puncher.Session()) {
		return fmt.Errorf("session mismatch in %s", expected)
	}
	return nil
}
