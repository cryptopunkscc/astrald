package objects

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"reflect"
)

var _ objects.Module = &Module{}

type Module struct {
	node   node2.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context

	prototypes sig.Map[string, desc.Data]
	openers    sig.Set[*Opener]
	creators   sig.Set[*Creator]
	describers sig.Set[objects.Describer]
	searchers  sig.Set[objects.Searcher]
	purgers    sig.Set[objects.Purger]
	finders    sig.Set[objects.Finder]
	decoders   sig.Map[string, objects.Decoder]

	provider *Provider

	content content.Module
	nodes   nodes.Module
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

func (mod *Module) Store(ctx context.Context, obj objects.Object) (objectID object.ID, err error) {
	var buf = &bytes.Buffer{}
	err = cslq.Encode(buf, "vv", adc.Header(obj.ObjectType()), obj)
	if err != nil {
		return
	}

	return mod.Put(buf.Bytes(), nil)
}

func (mod *Module) Load(ctx context.Context, objectID object.ID, scope *net.Scope) (objects.Object, error) {
	if objectID.Size > objects.ReadAllMaxSize {
		return nil, objects.ErrObjectTooLarge
	}

	r, err := mod.Open(ctx, objectID, &objects.OpenOpts{
		Zone:        scope.Zone,
		QueryFilter: scope.QueryFilter,
	})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	realID, obj, err := mod.decodeStream(r)
	if err != nil {
		return nil, err
	}
	if !realID.IsEqual(objectID) {
		return nil, objects.ErrHashMismatch
	}

	return obj, err
}

func (mod *Module) Get(id object.ID, opts *objects.OpenOpts) ([]byte, error) {
	if id.Size > objects.ReadAllMaxSize {
		return nil, errors.New("data too big")
	}
	r, err := mod.Open(context.Background(), id, opts)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	realID := object.Resolve(data)

	if !realID.IsEqual(id) {
		return nil, objects.ErrHashMismatch
	}

	return data, err
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
