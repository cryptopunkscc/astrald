package nat

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// NatSignal represents control messages exchanged over the signalling channel.
// Type field values are defined as constants below for readability and reuse.
type NatSignal struct {
	Type    astral.String `json:"type"`
	Session astral.Bytes  `json:"session"`
	IP      ip.IP         `json:"ip"`
	Port    astral.Uint16 `json:"port"`
}

const (
	NatSignalTypeOffer  = "offer"
	NatSignalTypeAnswer = "answer"
	NatSignalTypeReady  = "ready"
	NatSignalTypeGo     = "go"
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

func init() {
	_ = astral.DefaultBlueprints.Add(&NatSignal{})
}
