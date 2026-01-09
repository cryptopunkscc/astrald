package events

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Event struct {
	ID        astral.Nonce
	SourceID  *astral.Identity
	Timestamp astral.Time
	Data      astral.Object
}

func (Event) ObjectType() string {
	return "mod.events.event"
}

func (e Event) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *Event) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func init() {
	_ = astral.DefaultBlueprints.Add(&Event{})
}
