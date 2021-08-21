package push

import (
	"github.com/cryptopunkscc/astrald/services/push/internal/handle"
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/push/request"
)

func NewService() *service.Context {
	return service.New(Port, map[byte]service.Handle{
		request.Push:    handle.Push,
		request.Observe: handle.Observe,
	})
}
