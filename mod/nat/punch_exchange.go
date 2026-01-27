package nat

import (
	"bytes"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// PunchExchange coordinates the NAT punch signaling protocol over a channel.
type PunchExchange struct {
	ch      *channel.Channel
	Session []byte
}

func NewPunchExchange(ch *channel.Channel) *PunchExchange {
	return &PunchExchange{ch: ch}
}

func (e *PunchExchange) Send(sig *PunchSignal) error {
	sig.Session = e.Session
	return e.ch.Send(sig)
}

func (e *PunchExchange) Receive() (*PunchSignal, error) {
	obj, err := e.ch.Receive()
	if err != nil {
		return nil, err
	}

	switch sig := obj.(type) {
	case *astral.ErrorMessage:
		return nil, fmt.Errorf("received error: %s", sig.Error())
	case *PunchSignal:
		if e.Session != nil && !bytes.Equal(sig.Session, e.Session) {
			return nil, fmt.Errorf("session mismatch")
		}
		return sig, nil
	default:
		return nil, astral.NewErrUnexpectedObject(sig)
	}
}

func (e *PunchExchange) Expect(expected astral.String8) (*PunchSignal, error) {
	sig, err := e.Receive()
	if err != nil {
		return nil, err
	}
	if sig.Signal != expected {
		return nil, fmt.Errorf("unexpected signal: got %s, expected %s", sig.Signal, expected)
	}
	return sig, nil
}
