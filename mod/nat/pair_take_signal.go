package nat

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
)

const (
	PairHandoverSignalTypeLock     = "lock"
	PairHandoverSignalTypeLockOk   = "lock_ok"
	PairHandoverSignalTypeLockBusy = "lock_busy"
	PairHandoverSignalTypeTake     = "take"
	PairHandoverSignalTypeTakeOk   = "take_ok"
	PairHandoverSignalTypeTakeErr  = "take_err"
)

type PairTakeSignal struct {
	Signal astral.String8
	Pair   astral.Nonce
}

func (p PairTakeSignal) ObjectType() string { return "mod.nat.pair_take_signal" }

func (p PairTakeSignal) WriteTo(w io.Writer) (int64, error) { return astral.Struct(p).WriteTo(w) }

func (p *PairTakeSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(p).ReadFrom(r)
}

func (p PairTakeSignal) MarshalJSON() ([]byte, error) {
	type alias PairTakeSignal
	return json.Marshal(alias(p))
}

func (p *PairTakeSignal) UnmarshalJSON(bytes []byte) error {
	type alias PairTakeSignal
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*p = PairTakeSignal(a)
	return nil
}

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

func init() {
	_ = astral.Add(&PairTakeSignal{})
}
