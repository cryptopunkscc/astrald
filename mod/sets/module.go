package sets

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "sets"
const LocalNodeSet = "localnode"
const DBPrefix = "sets__"

type Module interface {
	CreateSet(set string, typ Type) (*Info, error)
	DeleteSet(set string) error
	AddToSet(set string, dataID data.ID) error
	RemoveFromSet(set string, dataID data.ID) error
	SetInfo(set string) (*Info, error)
	AllSets() ([]Info, error)
	Scan(name string, opts *ScanOpts) ([]*Member, error)
	Where(dataID data.ID) ([]string, error)
	Member(set string, dataID data.ID) (*Member, error)
	AddToUnion(superset string, subset string) error
	RemoveFromUnion(superset string, subset string) error

	SetVisible(set string, visible bool) error
	SetDescription(set string, desc string) error
}

type ScanOpts struct {
	UpdatedAfter   time.Time
	UpdatedBefore  time.Time
	IncludeRemoved bool
}

type Info struct {
	Name        string
	Type        Type
	Size        int
	Visible     bool
	Description string
	CreatedAt   time.Time
}

type Member struct {
	DataID    data.ID
	Added     bool
	UpdatedAt time.Time
}

type Type string

const (
	TypeSet   = Type("set")
	TypeUnion = Type("union")
)
