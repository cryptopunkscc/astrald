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
	db         *DB
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

func (mod *Module) Save(obj astral.Object) (_ *object.ID, err error) {
	realID, err := astral.ResolveObjectID(obj)
	if err != nil {
		return nil, err
	}

	if has, _ := mod.db.Contains(&realID); has {
		return &realID, nil
	}

	w, err := mod.Create(nil)
	if err != nil {
		return
	}
	defer w.Discard()

	_, err = astral.WriteCanonical(w, obj)

	if err != nil {
		return
	}

	emit := func(n bool) {
		mod.Objects.Receive(&objects.EventSaved{
			ObjectID: &realID,
			New:      astral.Bool(n),
		}, nil)
	}

	realID, err = w.Commit()
	switch {
	case err == nil:
		mod.onSave(&realID)
		emit(true)
		return &realID, nil

	case errors.Is(err, objects.ErrAlreadyExists):
		mod.onSave(&realID)
		emit(false)
		return &realID, nil
	}

	return
}

func (mod *Module) onSave(objectID *object.ID) {
	has, err := mod.db.Contains(objectID)
	switch {
	case err != nil:
		mod.log.Error("onSave: db error: %v", err)
		return
	case has:
		return
	}

	var ctx = astral.NewContext(nil).WithIdentity(mod.node.Identity())

	r, err := mod.Open(ctx, *objectID, nil)
	if err != nil {
		mod.log.Error("onSave: open %v error: %v", objectID, err)
		return
	}
	defer r.Close()

	var header astral.ObjectHeader
	_, err = header.ReadFrom(r)

	err = mod.db.Create(objectID, header.String())
	if err != nil {
		mod.log.Error("onSave: db error: %v", err)
	}
}

// Load loads and returns a typed object. Load verifies the hash of the loaded object.
func (mod *Module) Load(ctx *astral.Context, objectID *object.ID) (o astral.Object, err error) {
	if objectID.Size > uint64(MaxObjectSize) {
		return nil, errors.New("object too large")
	}

	r, err := mod.Open(ctx, *objectID, &objects.OpenOpts{
		Zone: astral.AllZones,
	})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	o, _, err = mod.Blueprints().Read(r, true)

	realID, err := astral.ResolveObjectID(o)
	if err != nil {
		return nil, errors.New("failed to load object")
	}

	if !realID.IsEqual(*objectID) {
		return nil, errors.New("failed to load object")
	}

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
