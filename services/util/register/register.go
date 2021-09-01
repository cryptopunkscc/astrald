package register

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
)

func Port(
	ctx context.Context,
	core api.Core,
	port string,
) (handler api.PortHandler, err error) {
	handler, err = core.Network().Register(port)
	if err != nil {
		return
	}
	go func() {
		<-ctx.Done()
		_ = handler.Close()
	}()
	return
}
