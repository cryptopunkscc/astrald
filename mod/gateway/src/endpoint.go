package gateway

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Endpoint = &Endpoint{}

type Endpoint struct {
	gate   *astral.Identity
	target *astral.Identity
}

// NewEndpoint insatntiates and returns a new Endpoint
func NewEndpoint(gate *astral.Identity, target *astral.Identity) *Endpoint {
	return &Endpoint{gate: gate, target: target}
}

// Pack returns a binary representation of the address
func (endpoint Endpoint) Pack() []byte {
	buf := &bytes.Buffer{}

	if err := cslq.Encode(buf, "vv", endpoint.gate, endpoint.target); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// Address returns a text representation of the address
func (endpoint Endpoint) Address() string {
	if endpoint.IsZero() {
		return "unknown"
	}
	return endpoint.gate.String() + ":" + endpoint.target.String()
}

func (endpoint Endpoint) IsZero() bool {
	return endpoint.gate.IsZero()
}

func (endpoint Endpoint) Gate() *astral.Identity {
	return endpoint.gate
}

func (endpoint Endpoint) Target() *astral.Identity {
	return endpoint.target
}

func (endpoint Endpoint) Network() string {
	return NetworkName
}

func (endpoint Endpoint) String() string {
	return endpoint.Network() + ":" + endpoint.Address()
}
