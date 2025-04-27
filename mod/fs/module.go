package fs

import (
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const (
	ModuleName = "fs"
	DBPrefix   = "fs__"
)

type Module interface {
	Find(opts *FindOpts) []*File
	Path(objectID *object.ID) []string
}

type File struct {
	Path     string
	ObjectID *object.ID
	ModTime  time.Time
}

type FindOpts struct {
	UpdatedAfter time.Time
}
