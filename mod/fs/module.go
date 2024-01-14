package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "fs"

type Module interface {
	Find(id data.ID) []string
}

type EventLocalFileChanged struct {
	Path      string
	OldID     data.ID
	NewID     data.ID
	IndexedAt time.Time
}
