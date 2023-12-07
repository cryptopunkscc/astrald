package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type API interface {
	Find(id data.ID) []string
}

type EventLocalFileChanged struct {
	Path      string
	OldID     data.ID
	NewID     data.ID
	IndexedAt time.Time
}
