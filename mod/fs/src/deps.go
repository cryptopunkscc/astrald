package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"time"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(fs.ModuleName, NewAdmin(mod))

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err := mod.Content.Ready(ctx); err != nil {
		return err
	}

	// add object blueprints
	mod.Objects.Blueprints().Add(&fs.FileDescriptor{})

	// add preconfigured repos
	for _, repo := range mod.repos.Clone() {
		mod.Objects.AddRepository(repo)
	}

	return
}
