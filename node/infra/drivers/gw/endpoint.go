package gw

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.Endpoint = Endpoint{}

type Endpoint struct {
	gate   id.Identity
	target id.Identity
}

// NewEndpoint insatntiates and returns a new Endpoint
func NewEndpoint(gate id.Identity, target id.Identity) Endpoint {
	return Endpoint{gate: gate, target: target}
}

// Pack returns a binary representation of the address
func (a Endpoint) Pack() []byte {
	buf := &bytes.Buffer{}

	if err := cslq.Encode(buf, "vv", a.gate, a.target); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// String returns a text representation of the address
func (a Endpoint) String() string {
	if a.IsZero() {
		return "unknown"
	}
	return a.gate.PublicKeyHex() + ":" + a.target.PublicKeyHex()
}

func (a Endpoint) IsZero() bool {
	return a.gate.IsZero()
}

func (a Endpoint) Gate() id.Identity {
	return a.gate
}

func (a Endpoint) Target() id.Identity {
	return a.target
}

func (a Endpoint) Network() string {
	return DriverName
}
