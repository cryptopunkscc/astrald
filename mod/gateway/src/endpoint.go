package gateway

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ exonet.Endpoint = &Endpoint{}
var _ astral.Object = &Endpoint{}

type Endpoint struct {
	gate   *astral.Identity
	target *astral.Identity
}

func (Endpoint) ObjectType() string {
	return "astrald.mod.gateway.endpoint"
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

func (endpoint Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	endpoint.gate = &astral.Identity{}
	endpoint.target = &astral.Identity{}
	return streams.ReadAllFrom(r, endpoint.gate, endpoint.target)
}

func (endpoint Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, endpoint.gate, endpoint.target)
}
