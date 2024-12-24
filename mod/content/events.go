package content

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &EventObjectIdentified{}

type EventObjectIdentified struct {
	TypeInfo *TypeInfo
}

func (EventObjectIdentified) ObjectType() string {
	return "astrald.mod.content.events.object_identified"
}

func (e EventObjectIdentified) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventObjectIdentified) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}
