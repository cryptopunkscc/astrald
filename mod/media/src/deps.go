package media

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

const IndexedSet = ".mod.media.index"

func (mod *Module) LoadDependencies() error {
	var err error

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(media.ModuleName, NewAdmin(mod))
	}

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.content.Ready(ctx); err != nil {
		return err
	}

	mod.content.AddDescriber(mod)
	mod.content.AddFinder(mod)
	mod.content.AddPrototypes(media.Desc{})

	// create our sets if needed
	mod.indexedSet, err = mod.sets.Open(IndexedSet, true)
	if err != nil {
		return err
	}

	return nil
}
