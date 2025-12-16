package nodes

import (
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type ObjectsModule struct{ t *testing.T }

var _ objects.Module = &ObjectsModule{}

func (o *ObjectsModule) Push(ctx *astral.Context, target *astral.Identity, obj astral.Object) error {
	o.t.Logf("%T.Push(%+v, %+v)", o, target, obj)
	return nil
}

func (o *ObjectsModule) AddRepository(id string, repo objects.Repository) error {
	panic("implement me")
}

func (o *ObjectsModule) Root() (repo objects.Repository) {
	panic("implement me")
}

func (o *ObjectsModule) AddDescriber(describer objects.Describer) error {
	panic("implement me")
}

func (o *ObjectsModule) Describe(context *astral.Context, id *astral.ObjectID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	panic("implement me")
}

func (o *ObjectsModule) AddPurger(purger objects.Purger) error {
	panic("implement me")
}

func (o *ObjectsModule) Purge(id *astral.ObjectID, opts *objects.PurgeOpts) (int, error) {
	panic("implement me")
}

func (o *ObjectsModule) Search(ctx *astral.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	panic("implement me")
}

func (o *ObjectsModule) AddSearcher(searcher objects.Searcher) error {
	panic("implement me")
}

func (o *ObjectsModule) AddSearchPreprocessor(preprocessor objects.SearchPreprocessor) error {
	panic("implement me")
}

func (o *ObjectsModule) AddFinder(finder objects.Finder) error {
	panic("implement me")
}

func (o *ObjectsModule) Find(context *astral.Context, id *astral.ObjectID) []*astral.Identity {
	panic("implement me")
}

func (o *ObjectsModule) AddHolder(holder objects.Holder) error {
	panic("implement me")
}

func (o *ObjectsModule) Holders(objectID *astral.ObjectID) []objects.Holder {
	panic("implement me")
}

func (o *ObjectsModule) AddReceiver(receiver objects.Receiver) error {
	panic("implement me")
}

func (o *ObjectsModule) Receive(object astral.Object, identity *astral.Identity) error {
	panic("implement me")
}

func (o *ObjectsModule) Blueprints() *astral.Blueprints {
	panic("implement me")
}

func (o *ObjectsModule) GetType(ctx *astral.Context, objectID *astral.ObjectID) (objectType string, err error) {
	panic("implement me")
}

func (o *ObjectsModule) On(target *astral.Identity, caller *astral.Identity) (objects.Consumer, error) {
	panic("implement me")
}
