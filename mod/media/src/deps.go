package media

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"time"
)

type Deps struct {
	Admin   admin.Module
	Auth    auth.Module
	Content content.Module
	Objects objects.Module
}

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

	mod.Objects.AddSearcher(mod)
	mod.Objects.AddOpener(mod, 20)
	mod.Objects.AddObject(&media.AudioDescriptor{})

	return
}
