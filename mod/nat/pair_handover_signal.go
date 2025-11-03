package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

const (
	PairHandoverSignalTypeLock     = "lock"
	PairHandoverSignalTypeLockOk   = "lock_ok"
	PairHandoverSignalTypeLockBusy = "lock_busy"
	PairHandoverSignalTypeTake     = "take"
	PairHandoverSignalTypeTakeOk   = "take_ok"
	PairHandoverSignalTypeTakeErr  = "take_err"
)

type PairHandoverSignal struct {
	Signal astral.String8
	PairID astral.Nonce
}

func (p PairHandoverSignal) ObjectType() string { return "mod.nat.pair_handover_signal" }

func (p PairHandoverSignal) WriteTo(w io.Writer) (int64, error) { return astral.Struct(p).WriteTo(w) }

func (p *PairHandoverSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(p).ReadFrom(r)
}

func (p PairHandoverSignal) MarshalJSON() ([]byte, error) {
	type alias PairHandoverSignal
	return json.Marshal(alias(p))
}

func (p *PairHandoverSignal) UnmarshalJSON(bytes []byte) error {
	type alias PairHandoverSignal
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*p = PairHandoverSignal(a)
	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&PairHandoverSignal{})
}
