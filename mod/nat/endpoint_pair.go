package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

// EndpointPair represents two peers that established a NAT-traversed
// connection.
type EndpointPair struct {
	PeerA     PeerEndpoint
	PeerB     PeerEndpoint
	CreatedAt astral.Time
}

// ObjectType implements astral.Object.
func (e EndpointPair) ObjectType() string {
	return "mod.nat.endpoint_pair"
}

// WriteTo implements astral.Object (binary serialization).
func (e EndpointPair) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(e).WriteTo(w)
}

// ReadFrom implements astral.Object (binary deserialization).
func (e *EndpointPair) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(e).ReadFrom(r)
}

// MarshalJSON encodes EndpointPair into JSON.
func (e EndpointPair) MarshalJSON() ([]byte, error) {
	type alias EndpointPair
	return json.Marshal(alias(e))
}

// UnmarshalJSON decodes EndpointPair from JSON.
func (e *EndpointPair) UnmarshalJSON(bytes []byte) error {
	type alias EndpointPair
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*e = EndpointPair(a)
	return nil
}

// PeerEndpoint represents a single peer's identity and endpoint.
// It can be serialized via astral.Struct and registered in blueprints.
type PeerEndpoint struct {
	Identity *astral.Identity
	Endpoint exonet.Endpoint
}

// ObjectType implements astral.Object.
func (p PeerEndpoint) ObjectType() string {
	return "mod.nat.peer_endpoint"
}

// WriteTo implements astral.Object (binary serialization).
func (p PeerEndpoint) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(p).WriteTo(w)
}

// ReadFrom implements astral.Object (binary deserialization).
func (p *PeerEndpoint) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(p).ReadFrom(r)
}

// MarshalJSON encodes PeerEndpoint into JSON.
func (p PeerEndpoint) MarshalJSON() ([]byte, error) {
	type alias PeerEndpoint
	return json.Marshal(alias(p))
}

// UnmarshalJSON decodes PeerEndpoint from JSON.
func (p *PeerEndpoint) UnmarshalJSON(bytes []byte) error {
	type alias PeerEndpoint
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*p = PeerEndpoint(a)
	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&PeerEndpoint{})
	_ = astral.DefaultBlueprints.Add(&EndpointPair{})
}
