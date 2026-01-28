package nat

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

const (
	PairTakeSignalTypeLock   = "lock"
	PairTakeSignalTypeLocked = "locked"
	PairTakeSignalTypeTake   = "take"
	PairTakeSignalTypeTaken  = "taken"
)

type PairTakeSignal struct {
	Signal astral.String8
	Pair   astral.Nonce
	Ok     bool
	Error  astral.String8
}

func (p PairTakeSignal) ObjectType() string { return "mod.nat.pair_take_signal" }

func (p PairTakeSignal) Err() error {
	if p.Error == "" {
		return errors.New("unknown error")
	}
	return errors.New(string(p.Error))
}

func (e PairTakeSignal) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *PairTakeSignal) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
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

func init() {
	_ = astral.Add(&PairTakeSignal{})
}
