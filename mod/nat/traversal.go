package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// Traversal is a pure state machine for NAT traversal protocol.
// All I/O operations (channel, puncher) are handled by callers.
type Traversal struct {
	Session []byte

	LocalIdentity *astral.Identity
	PeerIdentity  *astral.Identity

	LocalIP   ip.IP
	LocalPort astral.Uint16

	PeerIP   ip.IP
	PeerPort astral.Uint16

	Pair TraversedPortPair
}

func NewTraversal(localID, peerID *astral.Identity, localIP ip.IP) *Traversal {
	return &Traversal{
		LocalIdentity: localID,
		PeerIdentity:  peerID,
		LocalIP:       localIP,
	}
}

// OnOffer processes an incoming offer signal (participant side).
func (t *Traversal) OnOffer(sig *PunchSignal) {
	t.PeerIP = sig.IP
	t.PeerPort = sig.Port
	t.Session = sig.Session
}

// OnAnswer processes an incoming answer signal (initiator side).
func (t *Traversal) OnAnswer(sig *PunchSignal) {
	t.PeerIP = sig.IP
	t.PeerPort = sig.Port
}

// OnResult processes an incoming result signal.
func (t *Traversal) OnResult(sig *PunchSignal) {
	t.Pair.Nonce = sig.PairNonce
	t.Pair.PeerA.Endpoint = UDPEndpoint{
		IP:   sig.IP,
		Port: sig.Port,
	}
}

// ExpectSignal returns a handler for channel.Switch that validates the expected signal type.
func (t *Traversal) ExpectSignal(signalType astral.String8, on func(*PunchSignal)) func(*PunchSignal) error {
	return func(sig *PunchSignal) error {
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s, got %s", signalType, sig.Signal)
		}
		if on != nil {
			on(sig)
		}
		return channel.ErrBreak
	}
}

// SetPunchResult sets the pair from punch result.
func (t *Traversal) SetPunchResult(result *PunchResult) {
	t.Pair = TraversedPortPair{
		PeerA: PeerEndpoint{Identity: t.LocalIdentity},
		PeerB: PeerEndpoint{
			Identity: t.PeerIdentity,
			Endpoint: UDPEndpoint{IP: result.RemoteIP, Port: result.RemotePort},
		},
	}
}

// OfferSignal builds an offer signal to send to the peer.
func (t *Traversal) OfferSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeOffer,
		Session: t.Session,
		IP:      t.LocalIP,
		Port:    t.LocalPort,
	}
}

// AnswerSignal builds an answer signal to send to the peer.
func (t *Traversal) AnswerSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeAnswer,
		Session: t.Session,
		IP:      t.LocalIP,
		Port:    t.LocalPort,
	}
}

// ReadySignal builds a ready signal to send to the peer.
func (t *Traversal) ReadySignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeReady,
		Session: t.Session,
	}
}

// GoSignal builds a go signal to send to the peer.
func (t *Traversal) GoSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeGo,
		Session: t.Session,
	}
}

// ResultSignal builds a result signal to send to the peer.
func (t *Traversal) ResultSignal() *PunchSignal {
	return &PunchSignal{
		Signal:    PunchSignalTypeResult,
		Session:   t.Session,
		IP:        t.Pair.PeerB.Endpoint.IP,
		Port:      t.Pair.PeerB.Endpoint.Port,
		PairNonce: t.Pair.Nonce,
	}
}
