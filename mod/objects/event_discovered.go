package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type EventDiscovered struct {
	ObjectID *astral.ObjectID
	Zone     astral.Zone
}

// astral

func (EventDiscovered) ObjectType() string {
	return "mod.objects.events.discovered"
}

func (e EventDiscovered) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventDiscovered) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

// ...

func init() {
	_ = astral.Add(&EventDiscovered{})
}
