package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const ModuleName = "archives"
const DBPrefix = "archives__"

type Module interface {
	Index(context.Context, object.ID, *objects.OpenOpts) (*Archive, error)
	Forget(objectID object.ID) error
}

type Entry struct {
	ObjectID object.ID
	Path     string
	Comment  string
	Modified time.Time
}

type Archive struct {
	Entries []*Entry
	Comment string
	Format  string
}
