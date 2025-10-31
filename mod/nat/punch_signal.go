package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// PunchSignal represents control messages exchanged over the signalling channel.
// Signal field values are defined as constants below for readability and reuse.
type PunchSignal struct {
	Signal    astral.String8 `json:"signal"`
	Session   astral.Bytes8  `json:"session"`
	IP        ip.IP          `json:"ip"`
	Port      astral.Uint16  `json:"port"`
	PairNonce astral.Nonce   `json:"pair_nonce"`
}

const (
	PunchSignalTypeOffer  = "offer"
	PunchSignalTypeAnswer = "answer"
	PunchSignalTypeReady  = "ready"
	PunchSignalTypeGo     = "go"
	PunchSignalTypeResult = "result"
)

func (n PunchSignal) ObjectType() string {
	return "mod.nat.punch_signal"
}

func (n PunchSignal) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(n).WriteTo(w)
}

func (n *PunchSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(n).ReadFrom(r)
}

func (f PunchSignal) MarshalJSON() ([]byte, error) {
	type alias PunchSignal
	return json.Marshal(alias(f))
}

func (f *PunchSignal) UnmarshalJSON(bytes []byte) error {
	type alias PunchSignal
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*f = PunchSignal(a)
	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&PunchSignal{})
}
