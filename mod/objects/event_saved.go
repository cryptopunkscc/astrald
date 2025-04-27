package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type EventSaved struct {
	Identity *astral.Identity
	ObjectID *object.ID
	New      astral.Bool
}

var _ astral.Object = &EventSaved{}

func (e EventSaved) ObjectType() string {
	return "mod.objects.event_saved"
}

func (e EventSaved) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventSaved) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}
