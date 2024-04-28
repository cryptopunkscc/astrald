package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"time"
)

const ModuleName = "content"
const DBPrefix = "content__"
const IdentifiedSet = ".mod.content.identified"

type Module interface {
	Identify(dataID data.ID) (*TypeInfo, error)
	Forget(dataID data.ID) error
	Scan(ctx context.Context, opts *ScanOpts) <-chan *TypeInfo

	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	Finder
	AddFinder(Finder) error
	RemoveFinder(Finder) error

	AddPrototypes(protos ...desc.Data) error
	UnmarshalDescriptor(name string, buf []byte) desc.Data

	BestTitle(dataID data.ID) string
	Ready(ctx context.Context) error
}

type Describer desc.Describer[data.ID]

type ScanOpts struct {
	Type  string
	After time.Time
}

type TypeInfo struct {
	DataID       data.ID
	Type         string // detected data type
	Method       string // method used to detect type (adc | mimetype)
	IdentifiedAt time.Time
}

type EventDataIdentified struct {
	TypeInfo *TypeInfo
}
