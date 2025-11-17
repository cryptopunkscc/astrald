package nat

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// TraversedEndpoints represents two peers that established a NAT-traversed
// connection.
type TraversedEndpoints struct {
	PeerA     PeerEndpoint
	PeerB     PeerEndpoint
	CreatedAt astral.Time
	Nonce     astral.Nonce
}

// ObjectType implements astral.Object.
func (e TraversedEndpoints) ObjectType() string {
	return "mod.nat.traversed_endpoints"
}

// WriteTo implements astral.Object (binary serialization).
func (e TraversedEndpoints) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(e).WriteTo(w)
}

// ReadFrom implements astral.Object (binary deserialization).
func (e *TraversedEndpoints) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(e).ReadFrom(r)
}

// MarshalJSON encodes TraversedEndpoints into JSON.
func (e TraversedEndpoints) MarshalJSON() ([]byte, error) {
	type alias TraversedEndpoints
	return json.Marshal(alias(e))
}

// UnmarshalJSON decodes TraversedEndpoints from JSON.
func (e *TraversedEndpoints) UnmarshalJSON(bytes []byte) error {
	type alias TraversedEndpoints
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*e = TraversedEndpoints(a)
	return nil
}

// RemoteEndpoint returns the endpoint of the other peer in the pair.
func (e *TraversedEndpoints) RemoteEndpoint(self *astral.Identity) (PeerEndpoint, bool) {
	switch {
	case e.PeerA.Identity != nil && e.PeerA.Identity.IsEqual(self):
		return e.PeerB, true
	case e.PeerB.Identity != nil && e.PeerB.Identity.IsEqual(self):
		return e.PeerA, true
	default:
		return PeerEndpoint{}, false
	}
}

// PeerEndpoint represents a single peer's identity and endpoint.
// It can be serialized via astral.Struct and registered in blueprints.
type PeerEndpoint struct {
	Identity *astral.Identity
	// NOTE: cannot use exonet.Endpoint for serialization reasons,
	// and there is lack of package/struct describing (
	//transport layer) addr. (for now utp is only supported UDP protocol)
	Endpoint UDPEndpoint
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
	_ = astral.DefaultBlueprints.Add(&TraversedEndpoints{})
}
