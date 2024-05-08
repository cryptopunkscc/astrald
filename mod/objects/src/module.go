package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"reflect"
)

var _ objects.Module = &Module{}

// ReadAllMaxSize is the limit on data size accepted by Get() (to avoid accidental OOM)
var ReadAllMaxSize uint64 = 1024 * 1024 * 1024

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context

	prototypes sig.Map[string, desc.Data]
	openers    sig.Set[*Opener]
	creators   sig.Set[*Creator]
	describers sig.Set[objects.Describer]
	finders    sig.Set[objects.Finder]
	purgers    sig.Set[objects.Purger]

	provider *Provider

	content content.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.provider = NewProvider(mod)
	err := mod.node.LocalRouter().AddRoute("objects.*", mod.provider)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var describers []desc.Describer[object.ID]

	for _, d := range mod.describers.Clone() {
		describers = append(describers, d)
	}

	return desc.Collect(ctx, objectID, opts, describers...)
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) Find(ctx context.Context, query string, opts *objects.FindOpts) ([]objects.Match, error) {
	var matches []objects.Match
	var errs []error

	if opts == nil {
		opts = objects.DefaultFindOpts()
	}

	for _, finder := range mod.finders.Clone() {
		m, err := finder.Find(ctx, query, opts)
		if err != nil {
			errs = append(errs, err)
		}
		matches = append(matches, m...)
	}

	return matches, nil
}

func (mod *Module) AddFinder(finder objects.Finder) error {
	return mod.finders.Add(finder)
}

func (mod *Module) Get(id object.ID, opts *objects.OpenOpts) ([]byte, error) {
	if id.Size > ReadAllMaxSize {
		return nil, errors.New("data too big")
	}
	r, err := mod.Open(context.Background(), id, opts)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

func (mod *Module) Put(bytes []byte, opts *objects.CreateOpts) (object.ID, error) {
	if opts == nil {
		opts = &objects.CreateOpts{Alloc: len(bytes)}
	}

	w, err := mod.Create(opts)
	if err != nil {
		return object.ID{}, err
	}
	defer w.Discard()

	_, err = w.Write(bytes)
	if err != nil {
		return object.ID{}, err
	}

	return w.Commit()
}

func (mod *Module) Connect(caller id.Identity, target id.Identity) (objects.Consumer, error) {
	if target.IsZero() {
		return nil, errors.New("target cannot be zero")
	}

	if caller.IsZero() {
		caller = mod.node.Identity()
	}

	if caller.IsEqual(target) {
		return nil, errors.New("caller cannot be the same as target")
	}

	return NewConsumer(mod, caller, target), nil
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}

func (mod *Module) AddPrototypes(protos ...desc.Data) error {
	for _, proto := range protos {
		mod.prototypes.Set(proto.Type(), proto)
	}
	return nil
}

func (mod *Module) UnmarshalDescriptor(name string, buf []byte) desc.Data {
	p, ok := mod.prototypes.Get(name)
	if !ok {
		return nil
	}
	var v = reflect.ValueOf(p)

	c := reflect.New(v.Type())

	err := json.Unmarshal(buf, c.Interface())
	if err != nil {
		panic(err)
	}

	return c.Elem().Interface().(desc.Data)
}
