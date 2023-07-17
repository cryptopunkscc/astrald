package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
)

type CoreRouter struct {
	Routers *net.SerialRouter
	Monitor *MonitoredRouter
}

func NewCoreRouter(routers []net.Router, log *log.Logger) *CoreRouter {
	var router = &CoreRouter{
		Routers: net.NewSerialRouter(routers...),
	}

	router.Monitor = NewMonitoredRouter(router.Routers, log)

	return router
}

func (router *CoreRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return router.Monitor.RouteQuery(ctx, query, caller)
}

func (router *CoreRouter) Conns() *ConnSet {
	return router.Monitor.Conns()
}
