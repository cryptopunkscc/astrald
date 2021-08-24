package lore

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

const Port = "lore"

const storyMimeType = "application/lore"

type service struct {
	observers  map[sio.ReadWriteCloser]string
	repository repo2.LocalRepository
	ctx        context.Context
	core       api.Core
}

func run(ctx context.Context, core api.Core) error {
	srv := service{
		observers:  map[sio.ReadWriteCloser]string{},
		repository: repo.NewRepoClient(ctx, core),
		ctx: ctx,
		core: core,
	}

	go srv.observeIdentifier()

	// Handle incoming connections
	handle.Requests(ctx, core, Port, auth.Local, srv.Observe)
	return nil
}
