package sets

import (
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "sets"
const LocalNodeSet = "localnode"
const UniverseSet = "universe"
const DBPrefix = "sets__"

type Module interface {
	Create(set string, typ Type) (Set, error)
	CreateBasic(name string, members ...data.ID) (Basic, error)
	CreateUnion(name string, members ...string) (Union, error)
	SetInfo(set string) (*Info, error)
	All() ([]Info, error)
	Where(dataID data.ID) ([]string, error)

	Universe() Union
	Localnode() Union

	SetVisible(set string, visible bool) error
	SetDescription(set string, desc string) error

	SetOpener(typ Type, opener Opener)
	GetOpener(typ Type) Opener
	Open(set string) (Set, error)
	Edit(set string) (Editor, error)
}

type Opener func(name string) (Set, error)

type Set interface {
	Scan(opts *ScanOpts) ([]*Member, error)
	Info() (*Info, error)
}

type Basic interface {
	Set
	Add(dataID ...data.ID) error
	Remove(dataID ...data.ID) error
}

type Union interface {
	Set
	Add(name ...string) error
	Remove(name ...string) error
}

type Editor interface {
	Scan(opts *ScanOpts) ([]*Member, error)
	Add(...data.ID) error
	AddByID(...uint) error
	Remove(...data.ID) error
	RemoveByID(...uint) error
	Delete() error
}

type ScanOpts struct {
	UpdatedAfter   time.Time
	UpdatedBefore  time.Time
	IncludeRemoved bool
	DataID         data.ID
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
	Removed   bool
	UpdatedAt time.Time
}

type Type string

const (
	TypeBasic = Type("basic")
	TypeUnion = Type("union")
)

func Open[T any](mod Module, name string) (T, error) {
	var set any
	var err error
	var zero T

	set, err = mod.Open(name)
	if err != nil {
		return zero, err
	}
	if t, ok := set.(T); ok {
		return t, nil
	}
	return zero, errors.New("typecast failed")
}
