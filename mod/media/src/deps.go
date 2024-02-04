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
	mod.content.AddPrototypes(media.Descriptor{})

	// create our sets if needed
	mod.allSet, err = sets.Open[sets.Basic](mod.sets, media.AllSet)
	if err != nil {
		mod.allSet, err = mod.sets.CreateBasic(media.AllSet)
		if err != nil {
			return err
		}
		mod.sets.SetVisible(media.AllSet, true)
		mod.sets.SetDescription(media.AllSet, "All media")
	}

	return nil
}
