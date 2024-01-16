package index

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "index"

type Module interface {
	CreateIndex(name string, typ Type) (*Info, error)
	DeleteIndex(name string) error
	AddToSet(name string, dataID data.ID) error
	RemoveFromSet(name string, dataID data.ID) error
	IndexInfo(name string) (*Info, error)
	UpdatedSince(name string, since time.Time) ([]Entry, error)
	Contains(name string, dataID data.ID) (bool, error)
}

type Info struct {
	Name      string
	Type      Type
	Size      int
	CreatedAt time.Time
}

type Entry struct {
	DataID    data.ID
	Added     bool
	UpdatedAt time.Time
}

type Type string

const (
	TypeSet = Type("set")
)
