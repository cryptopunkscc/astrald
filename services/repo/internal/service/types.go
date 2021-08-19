package service

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

type Context struct {
	request.Context
	repo.ReadWriteMapRepository
	handlers  map[byte]Handle
	authorize Authorize
}

type Request struct {
	Context
}

type Handle func(c *Request)

type Handlers map[byte]Handle

type Authorize func(core api.Core, conn api.ConnectionRequest) bool
