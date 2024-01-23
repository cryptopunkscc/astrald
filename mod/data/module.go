package data

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/node/events"
	"time"
)

const ModuleName = "data"
const IdentifiedDataIndexName = "mod.data.identified"

type Module interface {
	// Events returns module's event queue
	Events() *events.Queue

	Identify(dataID data.ID) error
	FindByType(typ string, since time.Time) ([]TypeInfo, error)
	SubscribeType(ctx context.Context, typ string, since time.Time) <-chan TypeInfo

	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	SetLabel(data.ID, string)
	GetLabel(data.ID) string

	Ready(ctx context.Context) error
}

type Describer interface {
	DescribeData(ctx context.Context, dataID data.ID, opts *DescribeOpts) []Descriptor
}

type DescribeOpts struct {
}

type Descriptor struct {
	Type string
	Data any
}

const TypeDescriptorType = "mod.data.type"

type TypeDescriptor struct {
	Method string
	Type   string
}

const LabelDescriptorType = "mod.data.label"

type LabelDescriptor struct {
	Label string
}

type TypeInfo struct {
	DataID    data.ID
	IndexedAt time.Time
	Header    string
	Type      string
}

type EventDataIdentified TypeInfo

var ErrAlreadyIndexed = errors.New("already indexed")
