package astral

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ infra.Addr = &Addr{}

type Addr struct {
	gate   id.Identity
	target id.Identity
}

func NewAddr(gate id.Identity, target id.Identity) *Addr {
	return &Addr{gate: gate, target: target}
}

func (a Addr) Network() string {
	return NetworkName
}

func (a Addr) String() string {
	if a.IsZero() {
		return "unknown"
	}
	return a.gate.PublicKeyHex()
}

func (a Addr) Pack() []byte {
	buf := &bytes.Buffer{}

	enc.WriteIdentity(buf, a.gate)
	enc.WriteIdentity(buf, a.target)

	return buf.Bytes()
}

func (a Addr) IsZero() bool {
	return a.gate.IsZero()
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (Addr, error) {
	r := bytes.NewReader(data)

	nodeID, err := enc.ReadIdentity(r)
	if err != nil {
		return Addr{}, err
	}

	target, err := enc.ReadIdentity(r)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		gate:   nodeID,
		target: target,
	}, nil
}
