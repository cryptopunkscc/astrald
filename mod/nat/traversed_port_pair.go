package nat

import (
	"encoding/json"
	"io"
	"net"

	"github.com/cryptopunkscc/astrald/astral"
)

// TraversedPortPair represents two peers that established a NAT-traversed
// connection.
type TraversedPortPair struct {
	PeerA     PeerEndpoint
	PeerB     PeerEndpoint
	CreatedAt astral.Time
	Nonce     astral.Nonce
}

// ObjectType implements astral.Object.
func (e TraversedPortPair) ObjectType() string {
	return "mod.nat.traversed_port_pair"
}

// WriteTo implements astral.Object (binary serialization).
func (e TraversedPortPair) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(e).WriteTo(w)
}

// ReadFrom implements astral.Object (binary deserialization).
func (e *TraversedPortPair) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(e).ReadFrom(r)
}

// MarshalJSON encodes TraversedPortPair into JSON.
func (e TraversedPortPair) MarshalJSON() ([]byte, error) {
	type alias TraversedPortPair
	return json.Marshal(alias(e))
}

// UnmarshalJSON decodes TraversedPortPair from JSON.
func (e *TraversedPortPair) UnmarshalJSON(bytes []byte) error {
	type alias TraversedPortPair
	var a alias
	if err := json.Unmarshal(bytes, &a); err != nil {
		return err
	}
	*e = TraversedPortPair(a)
	return nil
}

// RemoteEndpoint returns the endpoint of the other peer in the pair.
func (e *TraversedPortPair) RemoteEndpoint(self *astral.Identity) (PeerEndpoint, bool) {
	switch {
	case e.PeerA.Identity != nil && e.PeerA.Identity.IsEqual(self):
		return e.PeerB, true
	case e.PeerB.Identity != nil && e.PeerB.Identity.IsEqual(self):
		return e.PeerA, true
	default:
		return PeerEndpoint{}, false
	}
}

// GetLocalAddr returns the local UDP address for this pair
func (p *TraversedPortPair) GetLocalAddr(self *astral.Identity) *net.UDPAddr {
	var local PeerEndpoint
	if p.PeerA.Identity.IsEqual(self) {
		local = p.PeerA
	} else {
		local = p.PeerB
	}
	return &net.UDPAddr{
		IP:   net.ParseIP(local.Endpoint.HostString()),
		Port: int(local.Endpoint.Port),
	}
}

// GetRemoteAddr returns the remote UDP address for this pair
func (p *TraversedPortPair) GetRemoteAddr(self *astral.Identity) *net.UDPAddr {
	var remote PeerEndpoint
	if p.PeerA.Identity.IsEqual(self) {
		remote = p.PeerB
	} else {
		remote = p.PeerA
	}
	return &net.UDPAddr{
		IP:   net.ParseIP(remote.Endpoint.HostString()),
		Port: int(remote.Endpoint.Port),
	}
}

func (p *TraversedPortPair) MatchesPeer(peer *astral.Identity) bool {
	return p.PeerA.Identity.IsEqual(peer) || p.PeerB.Identity.IsEqual(peer)
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
	_ = astral.DefaultBlueprints.Add(&TraversedPortPair{})
}
