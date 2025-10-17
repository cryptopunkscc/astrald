package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// NatSignal represents control messages exchanged over the signalling channel.
// Signal field values are defined as constants below for readability and reuse.
type NatSignal struct {
	Signal  astral.String8 `json:"signal"`
	Session astral.Bytes8  `json:"session"`
	IP      ip.IP          `json:"ip"`
	Port    astral.Uint16  `json:"port"`
}

const (
	NatSignalTypeOffer  = "offer"
	NatSignalTypeAnswer = "answer"
	NatSignalTypeReady  = "ready"
	NatSignalTypeGo     = "go"
	NatSignalTypeResult = "result"
)

func (n NatSignal) ObjectType() string {
	return "mod.nat.nat_signal"
}

func (n NatSignal) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(n).WriteTo(w)
}

func (n *NatSignal) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(n).ReadFrom(r)
}

func (f NatSignal) MarshalJSON() ([]byte, error) {
	type alias NatSignal
	return json.Marshal(alias(f))
}

func (f *NatSignal) UnmarshalJSON(bytes []byte) error {
	type alias NatSignal
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*f = NatSignal(a)
	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&NatSignal{})
}
