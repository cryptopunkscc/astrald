package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

const ModuleName = "fs"

type API interface {
	Find(id data.ID) []string
}

type EventLocalFileChanged struct {
	Path      string
	OldID     data.ID
	NewID     data.ID
	IndexedAt time.Time
}

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
