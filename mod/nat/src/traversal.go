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

// Role defines which side of the traversal we are.
type Role int

const (
	RoleInitiator Role = iota
	RoleResponder
)

// State defines coarse phases of the handshake.
type State int

const (
	StateOfferExchange  State = iota // Offer/Answer
	StateReadyPhase                  // Ready/Go
	StatePunch                       // Hole punching
	StateResultExchange              // Result exchange
	StateDone
)

type stateFn func(*astral.Context) (state State, err error)

// traversal is a tiny state machine driving NatSignal exchange.
type traversal struct {
	role   Role
	ch     *astral.Channel
	ips    []ip.IP // use ips[0]
	selfID *astral.Identity
	peerID *astral.Identity

	// runtime
	puncher   nat.Puncher
	localPort int

	// cached signals/data
	offer      *nat.NatSignal
	answer     *nat.NatSignal
	goSignal   *nat.NatSignal
	peerResult *nat.NatSignal

	// final
	ep nat.EndpointPair
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
	lp, err := t.puncher.Open()
	if err != nil {
		return err
	}
	t.localPort = lp
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

func (t *traversal) handleOfferExchange(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		err = t.setupPuncher(nil)
		if err != nil {
			return state, err
		}
		err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeOffer, Session: t.puncher.Session(), IP: t.ips[0], Port: astral.Uint16(t.localPort)})
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

	// responder
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
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeAnswer, Session: t.puncher.Session(), IP: t.ips[0], Port: astral.Uint16(t.localPort)})
	if err != nil {
		return state, err
	}
	state = StateReadyPhase
	return state, err
}

func (t *traversal) handleReadyPhase(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeReady, Session: t.puncher.Session()})
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
	// responder
	ready, err := t.recv()
	if err != nil {
		return state, err
	}
	if err = t.verify(ready, nat.NatSignalTypeReady); err != nil {
		return 0, err
	}
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeGo, Session: t.puncher.Session()})
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
	// our observation of peer endpoint
	observedPeer := utp.Endpoint{IP: res.RemoteIP, Port: res.RemotePort}
	// build base pair with identities; fill endpoints accordingly
	pair := nat.EndpointPair{
		PeerA: nat.PeerEndpoint{Identity: t.selfID},
		PeerB: nat.PeerEndpoint{Identity: t.peerID, Endpoint: observedPeer},
	}
	t.ep = pair
	state = StateResultExchange
	return state, err
}

func (t *traversal) handleResultExchange(ctx *astral.Context) (state State, err error) {
	if t.role == RoleInitiator {
		// send our observation of peer (PeerB)
		err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeResult, Session: t.puncher.Session(), IP: t.ep.PeerB.Endpoint.IP, Port: t.ep.PeerB.Endpoint.Port})
		if err != nil {
			return state, err
		}
		// receive peer's reported view of us
		res, err := t.recv()
		if err != nil {
			return state, err
		}
		if err = t.verify(res, nat.NatSignalTypeResult); err != nil {
			return 0, err
		}
		// complete pair: our endpoint as seen by peer goes to PeerA
		t.ep.PeerA.Endpoint = utp.Endpoint{IP: res.IP, Port: res.Port}
		state = StateDone
		return state, err
	}

	// responder: receive first, then send
	res, err := t.recv()
	if err != nil {
		return state, err
	}

	err = t.verify(res, nat.NatSignalTypeResult)
	if err != nil {
		return 0, err
	}
	// complete pair from responder perspective: PeerA is self (our endpoint as reported by peer)
	t.ep.PeerA.Endpoint = utp.Endpoint{IP: res.IP, Port: res.Port}
	// Now send our observation of peer (PeerB)
	err = t.ch.Write(&nat.NatSignal{Signal: nat.NatSignalTypeResult, Session: t.puncher.Session(), IP: t.ep.PeerB.Endpoint.IP, Port: t.ep.PeerB.Endpoint.Port})
	if err != nil {
		return state, err
	}
	state = StateDone
	return state, err
}

// Run executes the FSM and returns the resulting traversed endpoint pair.
func (t *traversal) Run(ctx *astral.Context) (ep nat.EndpointPair, err error) {
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
			return ep, fmt.Errorf("invalid state")
		}
		next, hErr := h(ctx)
		if hErr != nil {
			err = hErr
			return ep, err
		}
		state = next
	}
	ep = t.ep
	ep.CreatedAt = astral.Time(time.Now())
	return ep, err
}
