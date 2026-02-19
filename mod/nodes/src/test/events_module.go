package nodes

import (
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/events"
)

type EventsModule struct{ *testing.T }

var _ events.Module = &EventsModule{}

func (e *EventsModule) Emit(data astral.Object) *events.Event {
	e.T.Logf("%T.Emit(%+v)", e, data)
	return nil
}
