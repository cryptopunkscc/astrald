package storage

import (
	"github.com/cryptopunkscc/astrald/node/events"
)

type API interface {
	Events() *events.Queue
	Access() AccessManager
	Data() DataManager
}
