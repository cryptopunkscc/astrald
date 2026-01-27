package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// PairTakeExchange coordinates the pair take protocol over a channel.
type PairTakeExchange struct {
	ch   *channel.Channel
	Pair astral.Nonce
}

func NewPairTakeExchange(ch *channel.Channel, pair astral.Nonce) *PairTakeExchange {
	return &PairTakeExchange{ch: ch, Pair: pair}
}

func (e *PairTakeExchange) Send(signal astral.String8) error {
	return e.ch.Send(&PairTakeSignal{Signal: signal, Pair: e.Pair})
}

func (e *PairTakeExchange) Receive(ctx *astral.Context) (*PairTakeSignal, error) {
	var sig *PairTakeSignal
	err := e.ch.Switch(
		channel.WithContext(ctx),
		func(msg *PairTakeSignal) error {
			if msg.Pair != e.Pair {
				return fmt.Errorf("mismatched pair id %v (expected %v)", msg.Pair, e.Pair)
			}
			sig = msg
			return channel.ErrBreak
		},
		func(msg *astral.ErrorMessage) error {
			return msg
		},
	)
	if err != nil {
		return nil, err
	}
	if sig == nil {
		return nil, fmt.Errorf("missing pair take signal")
	}
	return sig, nil
}

func (e *PairTakeExchange) Expect(ctx *astral.Context, expected astral.String8) error {
	sig, err := e.Receive(ctx)
	if err != nil {
		return err
	}
	if sig.Signal != expected {
		return fmt.Errorf("unexpected signal: got %s, expected %s", sig.Signal, expected)
	}
	return nil
}

func (e *PairTakeExchange) SendReceive(ctx *astral.Context, signal astral.String8) (*PairTakeSignal, error) {
	if err := e.Send(signal); err != nil {
		return nil, err
	}
	return e.Receive(ctx)
}
