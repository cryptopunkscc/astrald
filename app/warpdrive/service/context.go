package service

import (
	"context"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"log"
)

// Context contains dependencies required by Handlers.
type Context struct {
	Config
	context.Context
	astral.Api
	api.Core
	*log.Logger
	Identity string
}

type Config struct {
	RepositoryDir  string
	RemoteResolver bool
}
