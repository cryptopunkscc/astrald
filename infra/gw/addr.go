package gw

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"strings"
)

var _ infra.Addr = Addr{}

type Addr struct {
	gate   id.Identity
	target id.Identity
}

// NewAddr insatntiates and returns a new Addr
func NewAddr(gate id.Identity, target id.Identity) Addr {
	return Addr{gate: gate, target: target}
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (addr Addr, err error) {
	return addr, cslq.Decode(bytes.NewReader(data), "vv", &addr.gate, &addr.target)
}

// Parse converts a text representation of a gateway address to an Addr struct
func Parse(str string) (addr Addr, err error) {
	if len(str) != (2*66)+1 { // two public key hex strings and a separator ":"
		return addr, errors.New("invalid address length")
	}
	var ids = strings.SplitN(str, ":", 2)
	if len(ids) != 2 {
		return addr, errors.New("invalid address string")
	}
	addr.gate, err = id.ParsePublicKeyHex(ids[0])
	if err != nil {
		return
	}
	addr.target, err = id.ParsePublicKeyHex(ids[1])
	return
}

// Pack returns a binary representation of the address
func (a Addr) Pack() []byte {
	buf := &bytes.Buffer{}

	if err := cslq.Encode(buf, "vv", a.gate, a.target); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// String returns a text representation of the address
func (a Addr) String() string {
	if a.IsZero() {
		return "unknown"
	}
	return a.gate.PublicKeyHex() + ":" + a.target.PublicKeyHex()
}

func (a Addr) IsZero() bool {
	return a.gate.IsZero()
}

func (a Addr) Gate() id.Identity {
	return a.gate
}

func (a Addr) Target() id.Identity {
	return a.target
}

func (a Addr) Network() string {
	return NetworkName
}
