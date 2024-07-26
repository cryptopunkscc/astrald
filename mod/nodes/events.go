package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"io"
)

var _ astral.Object = &EventLinked{}

type EventLinked struct {
	NodeID id.Identity
}

func (e EventLinked) ObjectType() string { return "mod.nodes.events.linked" }

func (e EventLinked) WriteTo(w io.Writer) (n int64, err error) {
	return e.NodeID.WriteTo(w)
}

func (e *EventLinked) ReadFrom(r io.Reader) (n int64, err error) {
	return e.NodeID.ReadFrom(r)
}

type EventUnlinked struct {
	NodeID id.Identity
}

func (e EventUnlinked) ObjectType() string { return "mod.nodes.events.unlinked" }

func (e EventUnlinked) WriteTo(w io.Writer) (n int64, err error) {
	return e.NodeID.WriteTo(w)
}

func (e *EventUnlinked) ReadFrom(r io.Reader) (n int64, err error) {
	return e.NodeID.ReadFrom(r)
}
