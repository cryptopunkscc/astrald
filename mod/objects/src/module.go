package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
)

var _ objects.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Auth    auth.Module
	Content content.Module
	Dir     dir.Module
	Nodes   nodes.Module
	Objects objects.Module
}

type Module struct {
	Deps
	blueprints astral.Blueprints
	node       astral.Node
	config     Config
	db         *gorm.DB
	log        *log.Logger
	ops        shell.Scope

	ctx        context.Context
	openers    sig.Set[*Opener]
	repos      sig.Map[string, objects.Repository]
	describers sig.Set[objects.Describer]
	searchers  sig.Set[objects.Searcher]
	purgers    sig.Set[objects.Purger]
	finders    sig.Set[objects.Finder]
	receivers  sig.Set[objects.Receiver]

	holders sig.Set[objects.Holder]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) Blueprints() *astral.Blueprints {
	return &mod.blueprints
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

func (mod *Module) Load(objectID object.ID) (o astral.Object, err error) {
	r, err := mod.Open(context.Background(), objectID, &objects.OpenOpts{
		Zone: astral.ZoneDevice | astral.ZoneVirtual,
	})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	o, _, err = mod.Blueprints().Read(r, true)

	return
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

func (mod *Module) Repositories() []objects.Repository {
	return mod.repos.Values()
}

func (mod *Module) On(target *astral.Identity, caller *astral.Identity) (objects.Consumer, error) {
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

func (mod *Module) String() string {
	return objects.ModuleName
}
