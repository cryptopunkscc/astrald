package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

var _ astral.Object = &EventCommitted{}

type EventCommitted struct {
	ObjectID *object.ID
}

func (EventCommitted) ObjectType() string {
	return "astrald.mod.objects.events.committed"
}

func (e EventCommitted) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventCommitted) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

var _ astral.Object = &EventDiscovered{}

type EventDiscovered struct {
	ObjectID *object.ID
	Zone     astral.Zone
}

func (EventDiscovered) ObjectType() string {
	return "astrald.mod.objects.events.discovered"
}

func (e EventDiscovered) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventDiscovered) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}
