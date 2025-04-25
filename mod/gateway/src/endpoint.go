package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"io"
	"strings"
)

type Endpoint struct {
	GatewayID *astral.Identity
	TargetID  *astral.Identity
}

var _ exonet.Endpoint = &Endpoint{}

func (Endpoint) ObjectType() string { return "mod.gateway.endpoint" }

func (e Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

// NewEndpoint makes a new Endpoint
func NewEndpoint(gateway *astral.Identity, target *astral.Identity) *Endpoint {
	return &Endpoint{GatewayID: gateway, TargetID: target}
}

// exonet.Endpoint

func (e Endpoint) Network() string {
	return NetworkName
}

// Address returns a text representation of the address
func (e Endpoint) Address() string {
	return e.GatewayID.String() + ":" + e.TargetID.String()
}

// Pack returns a binary representation of the address
func (e Endpoint) Pack() []byte {
	buf := &bytes.Buffer{}

	_, err := e.WriteTo(buf)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// text support

func (e *Endpoint) UnmarshalText(text []byte) (err error) {
	ids := strings.SplitN(string(text), ":", 2)
	if len(ids) != 2 {
		return errors.New("malformed endpoint")
	}

	e.GatewayID, err = astral.IdentityFromString(ids[0])
	if err != nil {
		return
	}

	e.TargetID, err = astral.IdentityFromString(ids[1])

	return
}

func (e Endpoint) MarshalText() (text []byte, err error) {
	return []byte(e.Address()), nil
}

// json support

func (e *Endpoint) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	return e.UnmarshalText([]byte(str))
}

func (e Endpoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Address())
}

// other

func (e Endpoint) IsZero() bool {
	return e.GatewayID.IsZero() && e.TargetID.IsZero()
}

func (e Endpoint) String() string {
	return e.Network() + ":" + e.Address()
}

func init() {
	astral.DefaultBlueprints.Add(&Endpoint{})
}
