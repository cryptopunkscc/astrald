package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type EventCommitted struct {
	ObjectID *astral.ObjectID
}

// astral

func (EventCommitted) ObjectType() string {
	return "mod.objects.events.committed"
}

func (e EventCommitted) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventCommitted) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&EventCommitted{})
}
