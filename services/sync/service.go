package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const Port = "sync"

const (
	Download = 1
)

func runService(ctx context.Context, core api.Core) error {
	rc := requestContext{
		Context:         ctx,
		Core:            core,
		LocalRepository: repo.NewRepoClient(ctx, core),
	}
	handlers := request.Handlers{
		Download: rc.Download,
	}
	go rc.syncLoop()
	handle.Requests(ctx, core, Port, auth.All, handle.Using(handlers))
	return nil
}
