package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"reflect"
)

var _ objects.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Auth    auth.Module
	Content content.Module
	Dir     dir.Module
	Nodes   nodes.Module
}

type Module struct {
	Deps
	*routers.PathRouter
	node   astral.Node
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
	objects    sig.Map[string, astral.Object]
	receivers  sig.Set[objects.Receiver]
	holders    sig.Set[objects.Holder]

	provider *Provider
}

func (mod *Module) AddObject(a astral.Object) error {
	_, ok := mod.objects.Set(a.ObjectType(), a)
	if !ok {
		return errors.New("object already added")
	}
	return nil
}

func (mod *Module) ReadObject(r io.Reader) (o astral.Object, err error) {
	var h astral.ObjectHeader
	_, err = h.ReadFrom(r)
	if err != nil {
		return
	}

	o = mod.getObject(h.String())
	if o == nil {
		o = &objects.ForeignObject{Type: h.String()}
	}

	_, err = o.ReadFrom(r)

	return
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.provider = NewProvider(mod)
	err := mod.AddRoute("objects.*", mod.provider)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) Store(obj astral.Object) (objectID object.ID, err error) {
	w, err := mod.Create(nil)
	if err != nil {
		return
	}
	defer w.Discard()

	_, err = astral.ObjectHeader(obj.ObjectType()).WriteTo(w)
	if err != nil {
		return
	}

	_, err = obj.WriteTo(w)

	return w.Commit()
}

func (mod *Module) Load(objectID object.ID) (astral.Object, error) {
	r, err := mod.Open(context.Background(), objectID, &objects.OpenOpts{
		Zone: astral.ZoneDevice | astral.ZoneVirtual,
	})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return mod.ReadObject(r)
}

func (mod *Module) Get(id object.ID, opts *objects.OpenOpts) ([]byte, error) {
	if id.Size > objects.ReadAllMaxSize {
		return nil, errors.New("data too big")
	}
	r, err := mod.Open(context.Background(), id, opts)
	if err != nil {
		return nil, err
	}

	res := object.NewReadResolver(r)

	data, err := io.ReadAll(res)
	if err != nil {
		return nil, err
	}

	if !res.Resolve().IsEqual(id) {
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

func (mod *Module) Connect(caller *astral.Identity, target *astral.Identity) (objects.Consumer, error) {
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

func (mod *Module) getObject(name string) astral.Object {
	p, ok := mod.objects.Get(name)
	if !ok {
		return nil
	}
	var v = reflect.ValueOf(p)
	var c = reflect.New(v.Elem().Type())

	return c.Interface().(astral.Object)
}
