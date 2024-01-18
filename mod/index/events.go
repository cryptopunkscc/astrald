package index

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type EventEntryUpdate struct {
	IndexName string
	DataID    data.ID
	Added     bool
	UpdatedAt time.Time
}

func (event EventEntryUpdate) String() string {
	if event.Added {
		return fmt.Sprintf("%s added %s", event.IndexName, event.DataID.String())
	}
	return fmt.Sprintf("%s removed %s", event.IndexName, event.DataID.String())
}

type EventIndexCreated struct {
	Info *Info
}

type EventIndexDeleted struct {
	Name string
}
