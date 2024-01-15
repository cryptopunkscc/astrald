package data

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"time"
)

const ModuleName = "data"

type Module interface {
	// Events returns module's event queue
	Events() *events.Queue

	FindByType(typ string, since time.Time) ([]TypeInfo, error)
	SubscribeType(ctx context.Context, typ string, since time.Time) <-chan TypeInfo

	OpenADC0(data.ID) (string, storage.DataReader, error)
	StoreADC0(t string, alloc int) (storage.DataWriter, error)

	Describer
	AddDescriber(Describer) error
	RemoveDescriber(Describer) error

	SetLabel(data.ID, string)
	GetLabel(data.ID) string
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

type EventDataIndexed TypeInfo
