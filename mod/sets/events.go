package sets

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type EventEntryUpdate struct {
	SetName   string
	DataID    data.ID
	Added     bool
	UpdatedAt time.Time
}

func (event EventEntryUpdate) String() string {
	if event.Added {
		return fmt.Sprintf("%s added %s", event.SetName, event.DataID.String())
	}
	return fmt.Sprintf("%s removed %s", event.SetName, event.DataID.String())
}

type EventSetCreated struct {
	Info *Info
}

type EventSetDeleted struct {
	Name string
}
