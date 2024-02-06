package sets

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "sets"
const DeviceSet = "device"
const VirtualSet = "virtual"
const NetworkSet = "network"
const UniverseSet = "universe"
const DBPrefix = "sets__"

type Module interface {
	Create(name string) (Set, error)
	CreateUnion(name string) (Union, error)
	CreateManaged(name string, typ Type) (Set, error)

	Open(name string, create bool) (Set, error)
	OpenUnion(name string, create bool) (Union, error)

	All() ([]string, error)
	Where(dataID data.ID) ([]string, error)

	SetWrapper(typ Type, fn WrapperFunc)
	Wrapper(typ Type) WrapperFunc

	Universe() Union
	Device() Union
	Virtual() Union
	Network() Union
}

type WrapperFunc func(Set) (Set, error)

type Set interface {
	Name() string
	DisplayName() string
	SetDisplayName(s string) error
	Scan(opts *ScanOpts) ([]*Member, error)
	Add(...data.ID) error
	AddByID(...uint) error
	Remove(...data.ID) error
	RemoveByID(...uint) error
	Delete() error
	Clear() error
	Trim(time.Time) error
	Stat() (*Stat, error)
}

type Union interface {
	Set
	Sync() error
	AddSubset(name ...string) error
	RemoveSubset(name ...string) error
	Subsets() ([]string, error)
}

type ScanOpts struct {
	UpdatedAfter   time.Time
	UpdatedBefore  time.Time
	IncludeRemoved bool
	DataID         data.ID
}

type Stat struct {
	Name        string
	Type        Type
	Size        int
	DataSize    uint64
	Visible     bool
	Description string
	CreatedAt   time.Time
	TrimmedAt   time.Time
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
