package handler

import (
	"context"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

// Context contains dependencies required by Handlers.
type Context struct {
	context.Context
	astral.Api
	api.Core
	Identity string
}

type Handler func(srv Context, request astral.Request)

type Handlers []map[string]Handler
