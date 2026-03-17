package nat

import (
	"bytes"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// PunchProtocol is a pure state machine for NAT traversal protocol.
type PunchProtocol struct {
	Session []byte

	LocalIdentity *astral.Identity
	PeerIdentity  *astral.Identity

	LocalIP   ip.IP
	LocalPort astral.Uint16

	PeerIP   ip.IP
	PeerPort astral.Uint16

	Hole Hole
}

func NewPunchProtocol(localID, peerID *astral.Identity, localIP ip.IP) *PunchProtocol {
	return &PunchProtocol{
		LocalIdentity: localID,
		PeerIdentity:  peerID,
		LocalIP:       localIP,
	}
}

func (t *PunchProtocol) OnOffer(sig *PunchSignal) {
	t.PeerIP = sig.IP
	t.PeerPort = sig.Port
	t.Session = sig.Session
}

func (t *PunchProtocol) OnAnswer(sig *PunchSignal) {
	t.PeerIP = sig.IP
	t.PeerPort = sig.Port
}

func (t *PunchProtocol) OnResult(sig *PunchSignal) {
	t.Hole.Nonce = sig.PairNonce
	t.Hole.ActiveEndpoint = Endpoint{IP: sig.IP, Port: sig.Port}
}

func (t *PunchProtocol) ExpectSignal(signalType astral.String8, on func(*PunchSignal)) func(*PunchSignal) error {
	return func(sig *PunchSignal) error {
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s, got %s", signalType, sig.Signal)
		}
		// Validate session for all signals except Offer (which establishes the session)
		if signalType != PunchSignalTypeOffer && !bytes.Equal(sig.Session, t.Session) {
			return fmt.Errorf("session mismatch")
		}
		if on != nil {
			on(sig)
		}
		return channel.ErrBreak
	}
}

func (t *PunchProtocol) SetPunchResult(result *PunchResult) {
	t.Hole = Hole{
		ActiveIdentity:  t.LocalIdentity,
		PassiveIdentity: t.PeerIdentity,
		PassiveEndpoint: Endpoint{IP: result.RemoteIP, Port: result.RemotePort},
	}
}

func (t *PunchProtocol) OfferSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeOffer,
		Session: t.Session,
		IP:      t.LocalIP,
		Port:    t.LocalPort,
	}
}

func (t *PunchProtocol) AnswerSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeAnswer,
		Session: t.Session,
		IP:      t.LocalIP,
		Port:    t.LocalPort,
	}
}

func (t *PunchProtocol) ReadySignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeReady,
		Session: t.Session,
	}
}

func (t *PunchProtocol) GoSignal() *PunchSignal {
	return &PunchSignal{
		Signal:  PunchSignalTypeGo,
		Session: t.Session,
	}
}

func (t *PunchProtocol) ResultSignal() *PunchSignal {
	return &PunchSignal{
		Signal:    PunchSignalTypeResult,
		Session:   t.Session,
		IP:        t.Hole.PassiveEndpoint.IP,
		Port:      t.Hole.PassiveEndpoint.Port,
		PairNonce: t.Hole.Nonce,
	}
}
