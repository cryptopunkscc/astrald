package media

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/resources"
)

type Module struct {
	Deps
	config Config
	node   astral.Node
	db     *DB
	log    *log.Logger
	assets resources.Resources
	ops    shell.Scope

	audio *AudioIndexer
	repo  *Repository
}

type Indexer interface {
	objects.Describer
	objects.Searcher
}

func (mod *Module) Run(ctx *astral.Context) error {
	go mod.indexer(ctx)

	mod.Objects.AddRepository("media-covers", mod.repo)

	<-ctx.Done()
	return nil
}

func (mod *Module) Index(ctx *astral.Context, objectID *object.ID) (err error) {
	// check if already indexed
	if _, e := mod.db.FindObject(objectID); e == nil {
		return nil
	}

	// mark as indexed
	defer mod.db.SaveObject(objectID)

	_, err = mod.audio.Index(ctx, objectID)
	switch {
	case errors.Is(err, objects.ErrNotFound):
		return err
	}

	return nil
}

func (mod *Module) Forget(ctx *astral.Context, objectID *object.ID) (err error) {
	mod.db.DeleteAudio(objectID)
	mod.db.DeleteObject(objectID)

	return nil
}

func (mod *Module) indexer(ctx *astral.Context) {
	ch, err := mod.Objects.Root().Scan(ctx, true)
	if err != nil {
		mod.log.Error("cannot scan objects: %v", err)
		return
	}

	for objectID := range ch {
		objectType, err := mod.Objects.GetType(ctx, objectID)

		switch {
		case err != nil:
			continue
		case objectType != "": // raw media formats have no object type
			continue
		}

		_ = mod.Index(ctx, objectID)
	}

	mod.log.Logv(1, "media indexer finished")
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return media.ModuleName
}
