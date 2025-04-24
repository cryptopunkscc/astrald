package gateway

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"io"
)

var _ exonet.Endpoint = &Endpoint{}
var _ astral.Object = &Endpoint{}

type Endpoint struct {
	GatewayID *astral.Identity
	TargetID  *astral.Identity
}

var _ astral.Object = &Endpoint{}

func (Endpoint) ObjectType() string {
	return "mod.gateway.endpoint"
}

func (endpoint Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(endpoint).WriteTo(w)
}

func (endpoint *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(endpoint).ReadFrom(r)
}

// NewEndpoint makes a new Endpoint
func NewEndpoint(gateway *astral.Identity, target *astral.Identity) *Endpoint {
	return &Endpoint{GatewayID: gateway, TargetID: target}
}

// Pack returns a binary representation of the address
func (endpoint Endpoint) Pack() []byte {
	buf := &bytes.Buffer{}

	_, err := endpoint.WriteTo(buf)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// Address returns a text representation of the address
func (endpoint Endpoint) Address() string {
	return endpoint.GatewayID.String() + ":" + endpoint.TargetID.String()
}

func (endpoint Endpoint) IsZero() bool {
	return endpoint.GatewayID.IsZero() && endpoint.TargetID.IsZero()
}

func (endpoint Endpoint) Network() string {
	return NetworkName
}

func (endpoint Endpoint) String() string {
	return endpoint.Network() + ":" + endpoint.Address()
}
