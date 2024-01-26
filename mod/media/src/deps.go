package media

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

func (mod *Module) LoadDependencies() error {
	var err error

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(media.ModuleName, NewAdmin(mod))
	}

	mod.data, err = modules.Load[data.Module](mod.node, data.ModuleName)
	if err != nil {
		return err
	}

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.index, err = modules.Load[index.Module](mod.node, index.ModuleName)
	if err != nil {
		return err
	}

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.data.Ready(ctx); err != nil {
		return err
	}

	mod.data.AddDescriber(mod)

	// create our indexes if needed
	if _, err = mod.index.IndexInfo(media.IndexNameAll); err != nil {
		_, err = mod.index.CreateIndex(media.IndexNameAll, index.TypeSet)
		if err != nil {
			return err
		}
		mod.index.SetVisible(media.IndexNameAll, true)
		mod.index.SetDescription(media.IndexNameAll, "All media")
	}

	return nil
}
