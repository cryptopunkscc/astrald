package fs

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
)

type EventFileChanged struct {
	Path  string
	OldID data.ID
	NewID data.ID
}

func (e EventFileChanged) String() string {
	return fmt.Sprintf("changed %s (%s -> %s)", e.Path, e.OldID, e.NewID)
}

type EventFileAdded struct {
	Path   string
	DataID data.ID
}

func (e EventFileAdded) String() string {
	return fmt.Sprintf("added %s (%s)", e.Path, e.DataID)
}

type EventFileRemoved struct {
	Path   string
	DataID data.ID
}

func (e EventFileRemoved) String() string {
	return fmt.Sprintf("removed %s (%s)", e.Path, e.DataID)
}
