package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

const (
	PairTakeSignalLock         = "lock"
	PairTakeSignalTypeLockOk   = "lock_ok"
	PairTakeSignalTypeLockBusy = "lock_busy"
	PairTakeSignalTypeTake     = "take"
	PairTakeSignalTypeTakeOk   = "take_ok"
	PairTakeSignalTypeTakeErr  = "take_err"
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

func init() {
	_ = astral.Add(&PairTakeSignal{})
}
