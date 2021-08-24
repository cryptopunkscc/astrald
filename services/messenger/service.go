package messenger

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/sio"
	repoService "github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const Port = "msg"

const (
	Send    = 1
	Observe = 2
)

type service struct {
	context.Context
	api.Core
	repo.LocalRepository
}

func run(ctx context.Context, core api.Core) error {
	observers := map[sio.ReadWriteCloser]struct{}{}
	srv := service{
		Context:         ctx,
		Core:            core,
		LocalRepository: repoService.NewRepoClient(ctx, core),
	}
	handlers := request.Handlers{
		Send: srv.Send,
		Observe: handle.Observe,
	}
	handler := handle.Using(handlers)

	go ObserveLore(ctx, core, Port, observers)
	handle.Requests(ctx, core, Port, auth.All, func(rc request.Context) error {
		rc.Observers = observers
		return handler(rc)
	})
	return nil
}
