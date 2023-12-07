package storage

import (
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "storage"

type API interface {
	Events() *events.Queue
	Access() AccessManager
	Data() DataManager
}

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
