package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const (
	ModuleName   = "objects"
	DBPrefix     = "objects__"
	ActionRead   = "objects.read"
	ActionCreate = "objects.create"
)

// MaxObjectSize is the maximum size of an object that can be loaded into memory
const MaxObjectSize int64 = 64 << 20 // 32 MB

type Module interface {
	// AddRepository registers a Repository
	AddRepository(id string, repo Repository) error
	Root() (repo Repository)

	AddDescriber(Describer) error
	Describe(*astral.Context, *object.ID, *astral.Scope) (<-chan *SourcedObject, error)

	AddPurger(purger Purger) error
	Purge(*object.ID, *PurgeOpts) (int, error)

	Search(ctx *astral.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)
	AddSearcher(Searcher) error
	AddSearchPreprocessor(SearchPreprocessor) error

	AddFinder(Finder) error
	Find(*astral.Context, *object.ID) []*astral.Identity

	AddHolder(Holder) error
	Holders(objectID *object.ID) []Holder

	AddReceiver(Receiver) error
	Receive(astral.Object, *astral.Identity) error

	Blueprints() *astral.Blueprints
	Push(ctx *astral.Context, target *astral.Identity, obj astral.Object) error
	GetType(ctx *astral.Context, objectID *object.ID) (objectType string, err error)

	// On returns a client for remote calls
	On(target *astral.Identity, caller *astral.Identity) (Consumer, error)
}

type Consumer interface {
	Describe(*astral.Context, *object.ID, *astral.Scope) (<-chan *SourcedObject, error)
	Search(*astral.Context, string) (<-chan *SearchResult, error)
	Push(*astral.Context, astral.Object) (err error)
}

type Receiver interface {
	ReceiveObject(Drop) error
}

type Drop interface {
	SenderID() *astral.Identity
	Object() astral.Object
	Accept(save bool) error
}

type Describer interface {
	DescribeObject(*astral.Context, *object.ID, *astral.Scope) (<-chan *SourcedObject, error)
}

type Purger interface {
	PurgeObject(*object.ID, *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}

type Holder interface {
	HoldObject(*object.ID) bool
}

func IsOffsetLimitValid(objectID *object.ID, offset int64, limit int64) bool {
	// offset has to be non-negative and cannot be larger than the object
	if offset < 0 || offset > int64(objectID.Size) {
		return false
	}

	// limit has to be non-negative and exceed the size of the object
	if limit < 0 || offset+limit > int64(objectID.Size) {
		return false
	}

	return true
}

func Save(ctx *astral.Context, object astral.Object, repo Repository) (objectID *object.ID, err error) {
	w, err := repo.Create(ctx, nil)
	if err != nil {
		return
	}
	defer w.Discard()

	_, err = astral.WriteCanonical(w, object)
	if err != nil {
		return
	}

	return w.Commit()
}
