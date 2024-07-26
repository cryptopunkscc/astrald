package media

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/media"
	"time"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(media.ModuleName, NewAdmin(mod))

	// wait for data module to finish preparing
	ctx, cancel := context.WithTimeoutCause(context.Background(), 15*time.Second, errors.New("data module timed out"))
	defer cancel()
	if err = mod.Content.Ready(ctx); err != nil {
		return err
	}

	mod.Objects.AddDescriber(mod)
	mod.Objects.AddSearcher(mod)
	mod.Objects.AddPrototypes(&media.Audio{})

	return
}
