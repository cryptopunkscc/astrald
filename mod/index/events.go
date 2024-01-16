package index

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type EventIndexEntryUpdate struct {
	IndexName string
	DataID    data.ID
	Added     bool
	UpdatedAt time.Time
}

type EventIndexCreated struct {
	Info *Info
}

type EventIndexDeleted struct {
	Name string
}
