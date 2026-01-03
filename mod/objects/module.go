package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
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
	Describe(*astral.Context, *astral.ObjectID) (<-chan *DescribeResult, error)

	AddPurger(purger Purger) error
	Purge(*astral.ObjectID, *PurgeOpts) (int, error)

	Search(ctx *astral.Context, query string, opts *SearchOpts) (<-chan *SearchResult, error)
	AddSearcher(Searcher) error
	AddSearchPreprocessor(SearchPreprocessor) error

	AddFinder(Finder) error
	Find(*astral.Context, *astral.ObjectID) []*astral.Identity

	AddHolder(Holder) error
	Holders(objectID *astral.ObjectID) []Holder

	AddReceiver(Receiver) error
	Receive(astral.Object, *astral.Identity) error

	Blueprints() *astral.Blueprints
	Push(ctx *astral.Context, target *astral.Identity, obj astral.Object) error
	GetType(ctx *astral.Context, objectID *astral.ObjectID) (objectType string, err error)
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
	DescribeObject(*astral.Context, *astral.ObjectID) (<-chan *DescribeResult, error)
}

type Purger interface {
	PurgeObject(*astral.ObjectID, *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}

type Holder interface {
	HoldObject(*astral.ObjectID) bool
}

func IsOffsetLimitValid(objectID *astral.ObjectID, offset int64, limit int64) bool {
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

func Save(ctx *astral.Context, object astral.Object, repo Repository) (objectID *astral.ObjectID, err error) {
	w, err := repo.Create(ctx, nil)
	if err != nil {
		return
	}
	defer w.Discard()

	_, err = astral.DefaultBlueprints.Canonical().Write(w, object)
	if err != nil {
		return
	}

	return w.Commit()
}
