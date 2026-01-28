package nat

import (
	"context"
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

// Traversal frames NAT traversal protocol messages over a channel.
type Traversal struct {
	ch      *channel.Channel
	session []byte

	LocalIdentity *astral.Identity
	PeerIdentity  *astral.Identity

	LocalIP   ip.IP
	LocalPort astral.Uint16

	PeerIP   ip.IP
	PeerPort astral.Uint16

	Puncher Puncher
	Pair    TraversedPortPair
}

func NewTraversal(ch *channel.Channel, localID, peerID *astral.Identity, localIP ip.IP, puncher Puncher) *Traversal {
	return &Traversal{
		ch:            ch,
		LocalIdentity: localID,
		PeerIdentity:  peerID,
		LocalIP:       localIP,
		Puncher:       puncher,
	}
}

func (t *Traversal) SendAnswer(ip ip.IP, port uint16, session []byte) error {
	t.session = session
	return t.ch.Send(&PunchSignal{
		Signal:  PunchSignalTypeAnswer,
		Session: session,
		IP:      ip,
		Port:    astral.Uint16(port),
	})
}

func (t *Traversal) SendGo() error {
	return t.ch.Send(&PunchSignal{
		Signal:  PunchSignalTypeGo,
		Session: t.session,
	})
}

func (t *Traversal) SendResult(ip ip.IP, port uint16, nonce astral.Nonce) error {
	return t.ch.Send(&PunchSignal{
		Signal:    PunchSignalTypeResult,
		Session:   t.session,
		IP:        ip,
		Port:      astral.Uint16(port),
		PairNonce: nonce,
	})
}

func (t *Traversal) SendOffer(ip ip.IP) error {
	localPort, err := t.Puncher.Open()
	if err != nil {
		return err
	}

	err = t.Puncher.SetSession([]byte(""))
	if err != nil {
		return err
	}

	t.LocalIP = ip

	return t.ch.Send(&PunchSignal{
		Signal:  PunchSignalTypeOffer,
		Session: t.session,
		IP:      ip,
		Port:    astral.Uint16(localPort),
	})
}

func (t *Traversal) SendReady() error {
	return t.ch.Send(&PunchSignal{
		Signal:  PunchSignalTypeReady,
		Session: t.session,
	})
}

func (t *Traversal) ExpectPunchSignal(signalType astral.String8, on func(sig *PunchSignal) error) func(sig *PunchSignal) error {
	return func(sig *PunchSignal) error {
		if sig.Signal != signalType {
			return fmt.Errorf("unexpected punch signal: %v", sig.Signal)
		}

		if on != nil {
			return on(sig)
		}

		return nil
	}
}

func (t *Traversal) OnOffer(sig *PunchSignal) error {
	if t.Puncher == nil {
		return errors.New("missing puncher")
	}
	if t.LocalIP == nil {
		return errors.New("missing local ip")
	}

	t.PeerIP = sig.IP
	t.PeerPort = sig.Port

	if err := t.Puncher.SetSession(sig.Session); err != nil {
		return err
	}

	if _, err := t.Puncher.Open(); err != nil {
		return err
	}

	return t.SendAnswer(
		t.LocalIP,
		uint16(t.Puncher.LocalPort()),
		t.Puncher.Session(),
	)
}

func (t *Traversal) OnAnswer(sig *PunchSignal) error {
	if t.Puncher == nil {
		return errors.New("missing puncher")
	}
	if t.LocalIP == nil {
		return errors.New("missing local ip")
	}

	t.PeerIP = sig.IP
	t.PeerPort = sig.Port
	return nil
}

func (t *Traversal) OnReady(ctx context.Context) func(_ *PunchSignal) error {
	return func(_ *PunchSignal) error {
		if t.Puncher == nil {
			return errors.New("missing puncher")
		}

		if t.LocalIdentity == nil || t.PeerIdentity == nil {
			return errors.New("missing identities")
		}

		if err := t.SendGo(); err != nil {
			return err
		}

		res, err := t.Puncher.HolePunch(ctx, t.PeerIP, int(t.PeerPort))
		if err != nil {
			return err
		}

		t.Pair = TraversedPortPair{
			PeerA: PeerEndpoint{
				Identity: t.LocalIdentity,
			},
			PeerB: PeerEndpoint{
				Identity: t.PeerIdentity,
				Endpoint: UDPEndpoint{
					IP:   res.RemoteIP,
					Port: res.RemotePort,
				},
			},
		}

		return t.SendResult(
			t.Pair.PeerB.Endpoint.IP,
			uint16(t.Pair.PeerB.Endpoint.Port),
			t.Pair.Nonce,
		)
	}
}

func (t *Traversal) OnResult(sig *PunchSignal) error {
	t.Pair.Nonce = sig.PairNonce
	t.Pair.PeerA.Endpoint = UDPEndpoint{
		IP:   sig.IP,
		Port: sig.Port,
	}

	return nil
}
