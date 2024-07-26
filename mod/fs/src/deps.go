package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/resources"
	"os"
	"path/filepath"
	"time"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Objects.AddOpener(mod, 30)
	mod.Objects.AddCreator(mod, 30)
	mod.Objects.AddDescriber(mod)
	mod.Objects.AddPurger(mod)
	mod.Objects.AddSearcher(NewFinder(mod))
	mod.Objects.AddPrototypes(fs.FileDesc{})

	mod.Admin.AddCommand(fs.ModuleName, NewAdmin(mod))

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.Content.Ready(ctx); err != nil {
		return err
	}

	// if we have file-based resources, use that as writable storage
	fileRes, ok := mod.assets.Res().(*resources.FileResources)
	if ok {
		mod.Watch(filepath.Join(fileRes.Root(), "static_data"))

		dataPath := filepath.Join(fileRes.Root(), "data")
		if os.MkdirAll(dataPath, 0700) == nil {
			mod.config.Store = append(mod.config.Store, dataPath)
			mod.Watch(dataPath)
		}
	}

	return
}
