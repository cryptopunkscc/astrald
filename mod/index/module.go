package index

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "index"
const LocalNodeUnionName = "localnode"

type Module interface {
	CreateIndex(name string, typ Type) (*Info, error)
	DeleteIndex(name string) error
	AddToSet(name string, dataID data.ID) error
	RemoveFromSet(name string, dataID data.ID) error
	IndexInfo(name string) (*Info, error)
	UpdatedBetween(name string, since time.Time, until time.Time) ([]Entry, error)
	Contains(name string, dataID data.ID) (bool, error)
	Find(dataID data.ID) ([]string, error)
	GetEntry(name string, dataID data.ID) (*Entry, error)
	AddToUnion(union string, set string) error
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
	TypeSet   = Type("set")
	TypeUnion = Type("union")
)
