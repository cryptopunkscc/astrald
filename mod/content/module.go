package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

const ModuleName = "content"
const DBPrefix = "content__"
const IdentifiedDataSetName = "mod.content.identified"

type Module interface {
	Identify(dataID data.ID) (*TypeInfo, error)
	IdentifySet(setName string) ([]*TypeInfo, error)
	Forget(dataID data.ID) error
	Scan(ctx context.Context, opts *ScanOpts) <-chan *TypeInfo

	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	AddPrototypes(protos ...DescriptorData) error
	UnmarshalDescriptor(name string, buf []byte) DescriptorData

	Ready(ctx context.Context) error
}

type ScanOpts struct {
	Type  string
	After time.Time
}

type Describer interface {
	Describe(ctx context.Context, dataID data.ID, opts *DescribeOpts) []*Descriptor
}

type DescribeOpts struct {
	Network        bool
	IdentityFilter func(id.Identity) bool
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
