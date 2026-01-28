package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// PairTakeExchange is a pure state machine for the pair take protocol.
// All I/O operations are handled by callers.
type PairTakeExchange struct {
	Pair astral.Nonce
}

func NewPairTakeExchange(pair astral.Nonce) *PairTakeExchange {
	return &PairTakeExchange{Pair: pair}
}

// Signal builders
func (e *PairTakeExchange) LockSignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeLock, Pair: e.Pair}
}

func (e *PairTakeExchange) LockOkSignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeLockOk, Pair: e.Pair}
}

func (e *PairTakeExchange) LockBusySignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeLockBusy, Pair: e.Pair}
}

func (e *PairTakeExchange) TakeSignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeTake, Pair: e.Pair}
}

func (e *PairTakeExchange) TakeOkSignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeTakeOk, Pair: e.Pair}
}

func (e *PairTakeExchange) TakeErrSignal() *PairTakeSignal {
	return &PairTakeSignal{Signal: PairTakeSignalTypeTakeErr, Pair: e.Pair}
}

// ExpectSignal returns a handler for channel.Switch that validates the expected signal type.
func (e *PairTakeExchange) ExpectSignal(signalType astral.String8) func(*PairTakeSignal) error {
	return func(sig *PairTakeSignal) error {
		if sig.Pair != e.Pair {
			return fmt.Errorf("mismatched pair id %v (expected %v)", sig.Pair, e.Pair)
		}
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s, got %s", signalType, sig.Signal)
		}
		return channel.ErrBreak
	}
}

// ReceiveSignal returns a handler that captures any valid signal for the pair.
func (e *PairTakeExchange) ReceiveSignal(out **PairTakeSignal) func(*PairTakeSignal) error {
	return func(sig *PairTakeSignal) error {
		if sig.Pair != e.Pair {
			return fmt.Errorf("mismatched pair id %v (expected %v)", sig.Pair, e.Pair)
		}
		*out = sig
		return channel.ErrBreak
	}
}
