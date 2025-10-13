package nat

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// NOTE: when astral.Slice will be improved we could just return astral.Slice

// TraversalResult represents the outcome of a NAT traversal negotiation.
// It holds both the address observed for the peer (by this node)
// and the address that the peer observed for us.
type TraversalResult struct {
	// PeerObserved is the external IP:port of the peer as seen by this node.
	PeerObservedIP   ip.IP
	PeerObservedPort astral.Uint16

	// SelfObserved is the external IP:port of this node as reported by the peer.
	ObservedIP   ip.IP
	ObservedPort astral.Uint16
}

func (n TraversalResult) ObjectType() string {
	return "mod.nat.nat_traversal_result"
}

func (n TraversalResult) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(n).WriteTo(w)
}

func (n *TraversalResult) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(n).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&TraversalResult{})
}
