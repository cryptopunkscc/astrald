package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "objects"

// ReadAllMaxSize is the size limit for loading objects into memory
const ReadAllMaxSize = 64 * 1024 * 1024 // 64 MB

type Module interface {
	// AddOpener registers an Opener. Openers are queried from highest to lowest priority.
	AddOpener(opener Opener, priority int) error
	Open(ctx context.Context, objectID object.ID, opts *OpenOpts) (Reader, error)

	// AddRepository registers a Repository
	AddRepository(repo Repository) error
	Repositories() []Repository
	Create(opts *CreateOpts) (Writer, error)

	AddDescriber(Describer) error
	Describe(context.Context, object.ID, *astral.Scope) (<-chan *SourcedObject, error)

	AddPurger(purger Purger) error
	Purge(object.ID, *PurgeOpts) (int, error)

	AddSearcher(Searcher) error
	Search(ctx context.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)

	AddFinder(Finder) error
	Find(context.Context, object.ID, *astral.Scope) []*astral.Identity

	AddHolder(Holder) error
	Holders(objectID object.ID) []Holder

	AddReceiver(Receiver) error
	Receive(astral.Object, *astral.Identity) error

	Blueprints() *astral.Blueprints
	Push(ctx context.Context, src *astral.Identity, dst *astral.Identity, obj astral.Object) error

	// Store encodes the object to local storage
	Store(astral.Object) (object.ID, error)
	Load(object.ID) (astral.Object, error)

	// Get reads the whole object into memory and returns the buffer
	Get(id object.ID, opts *OpenOpts) ([]byte, error)

	// Put commits the object to storage and returns its ID
	Put(object []byte, opts *CreateOpts) (object.ID, error)

	// On returns a client for remote calls
	On(target *astral.Identity, caller *astral.Identity) (Consumer, error)
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
	PurgeObject(object.ID, *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}

type Holder interface {
	HoldObject(object.ID) bool
}
