package objects

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
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

func (mod *Module) GetType(ctx *astral.Context, objectID *object.ID) (objectType string, err error) {
	row, err := mod.db.Find(objectID)
	if err == nil {
		return row.Type, nil
	}

	r, err := mod.Root().Read(ctx, objectID, 0, 260) // max header size: 4 magic bytes + 1 len + 255 type
	if err != nil {
		return "", objects.ErrNotFound
	}
	defer r.Close()

	var header astral.ObjectHeader
	_, err = header.ReadFrom(r)

	err = mod.db.Create(objectID, header.String())
	if err != nil {
		mod.log.Error("onSave: db error: %v", err)
	}

	return header.String(), nil
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
