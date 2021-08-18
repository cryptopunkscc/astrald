package service

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

type Service struct {
	port      string
	handlers  map[byte]Handle
	observers map[api.Stream]struct{}
	authorize Authorize
	repo      repo.ReadWriteRepository
}

type Handle func(c *request.Context)

type Handlers map[byte]Handle

type Authorize func(core api.Core, conn api.ConnectionRequest) bool
