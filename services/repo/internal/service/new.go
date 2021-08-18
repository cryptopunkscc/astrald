package service

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

func New(
	port string,
	authorize Authorize,
	handlers Handlers,
) *Context {
	return &Context{
		Context: request.Context{
			Port:      port,
			Observers: map[api.Stream]struct{}{},
		},
		authorize: authorize,
		handlers:  handlers,
	}
}
