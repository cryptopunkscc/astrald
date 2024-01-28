package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "content"
const IdentifiedDataSetName = "mod.content.identified"

type Module interface {
	Identify(dataID data.ID) (*Info, error)
	IdentifySet(setName string) ([]*Info, error)
	Forget(dataID data.ID) error
	Scan(ctx context.Context, opts *ScanOpts) <-chan *Info

	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	SetLabel(data.ID, string)
	GetLabel(data.ID) string

	Ready(ctx context.Context) error
}

type ScanOpts struct {
	Type  string
	After time.Time
}

type Describer interface {
	Describe(ctx context.Context, dataID data.ID, opts *DescribeOpts) []Descriptor
}

type DescribeOpts struct {
	// for future use
}

type Info struct {
	DataID    data.ID
	IndexedAt time.Time
	Method    string // method used to detect type (adc | mimetype)
	Type      string // detected data type
}

type EventDataIdentified struct {
	Info *Info
}
