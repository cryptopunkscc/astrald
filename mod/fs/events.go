package fs

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/object"
)

type EventFileChanged struct {
	Path  string
	OldID object.ID
	NewID object.ID
}

func (e EventFileChanged) String() string {
	return fmt.Sprintf("changed %s (%s -> %s)", e.Path, e.OldID, e.NewID)
}

type EventFileAdded struct {
	Path     string
	ObjectID object.ID
}

func (e EventFileAdded) String() string {
	return fmt.Sprintf("added %s (%s)", e.Path, e.ObjectID)
}

type EventFileRemoved struct {
	Path     string
	ObjectID object.ID
}

func (e EventFileRemoved) String() string {
	return fmt.Sprintf("removed %s (%s)", e.Path, e.ObjectID)
}
