package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

// ExpectPairTakeSignal returns a handler that validates pair and signal type.
func ExpectPairTakeSignal(pair astral.Nonce, signalType astral.String8, on func(*PairTakeSignal) error) func(*PairTakeSignal) error {
	return func(sig *PairTakeSignal) error {
		if sig.Pair != pair {
			return fmt.Errorf("mismatched pair id %v (expected %v)", sig.Pair, pair)
		}
		if sig.Signal != signalType {
			return fmt.Errorf("expected %s, got %s", signalType, sig.Signal)
		}
		if on != nil {
			if err := on(sig); err != nil {
				return err
			}
		}
		return channel.ErrBreak
	}
}

func HandleFailedPairTakeSignal(sig *PairTakeSignal) error {
	if !sig.Ok {
		return sig.Err()
	}

	return nil
}
