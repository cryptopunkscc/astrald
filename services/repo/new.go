package repo

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/repo/internal/adapter"
	"github.com/cryptopunkscc/astrald/services/repo/internal/auth"
	"github.com/cryptopunkscc/astrald/services/repo/internal/handle"
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"github.com/cryptopunkscc/astrald/services/repo/request"
)

func NewRepoClient(
	ctx context.Context,
	core api.Core,
) repo.LocalRepository {
	return adapter.New(Port, "", ctx, core)
}

func NewFilesClient(
	ctx context.Context,
	core api.Core,
	identity api.Identity,
) repo.RemoteRepository {
	return adapter.New(FilesPort, identity, ctx, core)
}

func NewRepoService() *service.Context {
	return service.New(Port,
		auth.AcceptLocal,
		service.Handlers{
			request.List:    handle.List,
			request.Read:    handle.Read,
			request.Write:   handle.Write,
			request.Observe: handle.Observe,
		},
	)
}

func NewFilesService() *service.Context {
	return service.New(FilesPort,
		auth.AcceptAll,
		service.Handlers{
			request.List: handle.List,
			request.Read: handle.Read,
		},
	)
}