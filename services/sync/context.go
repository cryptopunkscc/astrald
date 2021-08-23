package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
)

type requestContext struct {
	context.Context
	api.Core
	repo.LocalRepository
}