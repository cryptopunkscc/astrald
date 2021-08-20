package service

import (
	"github.com/cryptopunkscc/astrald/components/sio"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

func New(
	port string,
	authorize auth.Authorize,
	handlers Handlers,
) *Context {
	return &Context{
		Context: request.Context{
			Port:      port,
			Observers: map[sio.ReadWriteCloser]struct{}{},
		},
		authorize: authorize,
		handlers:  handlers,
	}
}
