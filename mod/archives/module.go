package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"time"
)

const ModuleName = "archives"
const DBPrefix = "archives__"

type Module interface {
	Index(context.Context, *astral.ObjectID) (*Archive, error)
	Forget(objectID *astral.ObjectID) error
}

type Entry struct {
	ObjectID *astral.ObjectID
	Path     string
	Comment  string
	Modified time.Time
}

type Archive struct {
	Entries []*Entry
	Comment string
	Format  string
}
