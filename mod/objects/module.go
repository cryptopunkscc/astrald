package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	ModuleName   = "objects"
	DBPrefix     = "objects__"
	ActionRead   = "objects.read"
	ActionCreate = "objects.create"

	RepoMain      = "main"      // everything
	RepoDevice    = "device"    // device: memory, local, removable
	RepoMemory    = "memory"    // memcache repos
	RepoLocal     = "local"     // local storage
	RepoRemovable = "removable" // removable storage
	RepoVirtual   = "virtual"   // virtual repos (archives, encryption, chunks)
	RepoNetwork   = "network"   // network repos
)

// MaxObjectSize is the maximum size of an object that can be loaded into memory
const MaxObjectSize int64 = 64 << 20 // 32 MB

type Module interface {
	// AddRepository registers a Repository
	AddRepository(name string, repo Repository) error

	// GetRepository returns a Repository by its name
	GetRepository(name string) Repository

	// ReadDefault returns the default repository for reading
	ReadDefault() Repository

	// WriteDefault returns the default repository for writing
	WriteDefault() Repository

	// AddGroup adds a repository to a group
	AddGroup(groupName string, repoName string) error

	// Load loads an object from a repository
	Load(*astral.Context, Repository, *astral.ObjectID) (astral.Object, error)

	// Store stores an object in a repository
	Store(*astral.Context, Repository, astral.Object) (*astral.ObjectID, error)

	AddDescriber(Describer) error
	Describe(*astral.Context, *astral.ObjectID) (<-chan *DescribeResult, error)

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
