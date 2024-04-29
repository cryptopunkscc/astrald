package sets

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type EventMemberUpdate struct {
	Set       string
	ObjectID  object.ID
	Removed   bool
	UpdatedAt time.Time
}

func (event EventMemberUpdate) String() string {
	if event.Removed {
		return fmt.Sprintf("%s added %s", event.Set, event.ObjectID.String())
	}
	return fmt.Sprintf("%s removed %s", event.Set, event.ObjectID.String())
}

type EventSetCreated struct {
	Set Set
}

type EventSetDeleted struct {
	Name string
}

type EventSetUpdated struct {
	Name string
}
