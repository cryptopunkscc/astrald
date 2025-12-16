package nodes

import (
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/events"
)

type EventsModule struct{ t *testing.T }

var _ events.Module = &EventsModule{}

func (e *EventsModule) Emit(data astral.Object) *events.Event {
	e.t.Logf("%T.Emit(%+v)", e, data)
	return nil
}
