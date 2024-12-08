package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
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
	Receive(astral.Object, *astral.Identity) error
	Push(ctx context.Context, src *astral.Identity, dst *astral.Identity, obj astral.Object) error

	// Store encodes the object to local storage
	Store(astral.Object) (object.ID, error)
	Load(object.ID) (astral.Object, error)

	// Get reads the whole object into memory and returns the buffer
	Get(id object.ID, opts *OpenOpts) ([]byte, error)

	// Put commits the object to storage and returns its ID
	Put(object []byte, opts *CreateOpts) (object.ID, error)

	Connect(target *astral.Identity, caller *astral.Identity) (Consumer, error)
}

type Consumer interface {
	Describe(context.Context, object.ID, *astral.Scope) (<-chan *SourcedObject, error)
	Open(context.Context, object.ID, *OpenOpts) (Reader, error)
	Put(context.Context, []byte) (object.ID, error)
	Search(context.Context, string) (<-chan *SearchResult, error)
	Push(context.Context, astral.Object) (err error)
}

type Receiver interface {
	ReceiveObject(*SourcedObject) error
}

type Describer interface {
	DescribeObject(context.Context, object.ID, *astral.Scope) (<-chan *SourcedObject, error)
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
