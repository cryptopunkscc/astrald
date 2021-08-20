package client

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
)

func New(
	port string,
	identity api.Identity,
	ctx context.Context,
	core api.Core,
) repo.LocalRepository {
	return &client{
		port:     port,
		identity: identity,
		ctx:      ctx,
		core:     core,
	}
}
