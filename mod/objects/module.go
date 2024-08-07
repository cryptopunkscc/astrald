package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

const ModuleName = "objects"

// ReadAllMaxSize is the size limit for loading objects into memory
const ReadAllMaxSize = 64 * 1024 * 1024 // 64 MB

type Module interface {
	// AddOpener registers an Opener. Openers are queried from highest to lowest priority.
	AddOpener(opener Opener, priority int) error
	Opener

	// AddCreator registers a Creator. Creators are queried from highest to lowest priority.
	AddCreator(creator Creator, priority int) error
	Creator

	AddDescriber(Describer) error
	Describer

	AddPurger(purger Purger) error
	Purger

	AddSearcher(Searcher) error
	Searcher

	AddFinder(Finder) error
	Finder

	AddHolder(Holder) error
	Holders(objectID object.ID) []Holder

	AddObject(astral.Object) error
	ReadObject(r io.Reader) (o astral.Object, err error)
	AddReceiver(Receiver) error
	Push(ctx context.Context, src *astral.Identity, dst *astral.Identity, obj astral.Object) error
	PushLocal(*astral.Identity, astral.Object) error

	// Store encodes the object to local storage
	Store(astral.Object) (object.ID, error)
	Load(object.ID) (astral.Object, error)

	AddPrototypes(protos ...desc.Data) error
	UnmarshalDescriptor(name string, buf []byte) desc.Data

	// Get reads the whole object into memory and returns the buffer
	Get(id object.ID, opts *OpenOpts) ([]byte, error)

	// Put commits the object to storage and returns its ID
	Put(object []byte, opts *CreateOpts) (object.ID, error)

	Connect(caller *astral.Identity, target *astral.Identity) (Consumer, error)
}

type Consumer interface {
	Describe(context.Context, object.ID, *desc.Opts) ([]*desc.Desc, error)
	Open(context.Context, object.ID, *OpenOpts) (Reader, error)
	Put(context.Context, []byte) (object.ID, error)
	Search(context.Context, string) ([]Match, error)
	Push(context.Context, astral.Object) (err error)
}

type Receiver interface {
	ReceiveObject(*SourcedObject) error
}

type SourcedObject struct {
	Source *astral.Identity
	Object astral.Object
}

type Describer interface {
	Describe(ctx context.Context, object object.ID, opts *desc.Opts) []*desc.Desc
}

type Purger interface {
	Purge(object.ID, *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}

type Holder interface {
	HoldObject(object.ID) bool
}
