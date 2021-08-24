package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const Port = "sync"

const (
	Download = 1
)

type service struct {
	context.Context
	api.Core
	repo2.LocalRepository
}

func run(ctx context.Context, core api.Core) error {
	srv := service{
		Context:         ctx,
		Core:            core,
		LocalRepository: repo.NewRepoClient(ctx, core),
	}
	handlers := request.Handlers{
		Download: srv.Download,
	}
	go srv.syncLoop()
	handle.Requests(ctx, core, Port, auth.All, handle.Using(handlers))
	return nil
}
