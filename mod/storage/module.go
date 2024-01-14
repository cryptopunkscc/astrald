package storage

import (
	"github.com/cryptopunkscc/astrald/node/events"
)

const ModuleName = "storage"

type Module interface {
	Events() *events.Queue
	Access() AccessManager
	Data() DataManager
}
