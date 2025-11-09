package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &NodeLinkedEvent{}

type NodeLinkedEvent struct {
	NodeID *astral.Identity
}

func (e NodeLinkedEvent) ObjectType() string { return "mod.nodes.events.linked" }

func (e NodeLinkedEvent) WriteTo(w io.Writer) (n int64, err error) {
	return e.NodeID.WriteTo(w)
}

func (e *NodeLinkedEvent) ReadFrom(r io.Reader) (n int64, err error) {
	return e.NodeID.ReadFrom(r)
}
