package events

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "events"

type Module interface {
	Emit(data astral.Object) *Event
}
