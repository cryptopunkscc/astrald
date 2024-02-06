package sets

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type EventMemberUpdate struct {
	Set       string
	DataID    data.ID
	Removed   bool
	UpdatedAt time.Time
}

func (event EventMemberUpdate) String() string {
	if event.Removed {
		return fmt.Sprintf("%s added %s", event.Set, event.DataID.String())
	}
	return fmt.Sprintf("%s removed %s", event.Set, event.DataID.String())
}

type EventSetCreated struct {
	Stat *Stat
}

type EventSetDeleted struct {
	Name string
}

type EventSetUpdated struct {
	Name string
}
