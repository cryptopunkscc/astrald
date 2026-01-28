package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// ExpectPairTakeSignal returns a handler that validates pair and signal type.
func ExpectPairTakeSignal(pair astral.Nonce, signalType astral.String8) func(*PairTakeSignal) error {
	return func(sig *PairTakeSignal) error {
		if sig.Pair != pair {
			return fmt.Errorf("mismatched pair id %v (expected %v)", sig.Pair, pair)
		}
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s, got %s", signalType, sig.Signal)
		}
		return channel.ErrBreak
	}
}

// ReceivePairTakeSignal returns a handler that validates pair and captures any signal.
func ReceivePairTakeSignal(pair astral.Nonce, out **PairTakeSignal) func(*PairTakeSignal) error {
	return func(sig *PairTakeSignal) error {
		if sig.Pair != pair {
			return fmt.Errorf("mismatched pair id %v (expected %v)", sig.Pair, pair)
		}
		*out = sig
		return channel.ErrBreak
	}
}
