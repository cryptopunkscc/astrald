package objects

import (
	"context"
	"errors"
	"fmt"
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
	describers sig.Set[objects.Describer]
	searchers  sig.Set[objects.Searcher]
	purgers    sig.Set[objects.Purger]
	finders    sig.Set[objects.Finder]
	receivers  sig.Set[objects.Receiver]
	holders    sig.Set[objects.Holder]
	repos      sig.Map[string, objects.Repository]
	root       *RootRepository
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

	r, err := mod.Root().Read(ctx, objectID, 0, 0)
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
	if objectID.Size > uint64(objects.MaxObjectSize) {
		return nil, errors.New("object too large")
	}

	r, err := mod.Root().Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	o, _, err = mod.Blueprints().Read(r, true)
	if err != nil {
		return nil, err
	}

	realID, err := astral.ResolveObjectID(o)
	if err != nil {
		return nil, errors.New("failed to load object")
	}

	if !realID.IsEqual(objectID) {
		return nil, errors.New("failed to load object")
	}

	return
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

func (mod *Module) AddRepository(id string, repo objects.Repository) error {
	_, ok := mod.repos.Set(id, repo)
	if !ok {
		return fmt.Errorf("repo %s already added", repo.Label())
	}
	return nil
}

func (mod *Module) GetRepository(id string) (repo objects.Repository, err error) {
	if id == "" {
		return mod.Root(), nil
	}

	repo, ok := mod.repos.Get(id)
	if !ok {
		err = errors.New("repository not found")
	}
	return
}

func (mod *Module) Root() (repo objects.Repository) {
	return mod.root
}

func (mod *Module) String() string {
	return objects.ModuleName
}
