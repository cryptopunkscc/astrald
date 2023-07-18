package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"time"
)

type CoreRouter struct {
	Routers *net.SerialRouter
	Monitor *MonitoredRouter
	log     *log.Logger
}

func NewCoreRouter(routers []net.Router, log *log.Logger) *CoreRouter {
	var router = &CoreRouter{
		Routers: net.NewSerialRouter(routers...),
		log:     log,
	}

	router.Monitor = NewMonitoredRouter(router.Routers)

	return router
}

func (router *CoreRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	router.log.Logv(2, "routing query %v -> %v:%v (origin %s)...", query.Caller(), query.Target(), query.Query(), query.Origin())

	var startedAt = time.Now()
	target, err := router.Monitor.RouteQuery(ctx, query, caller)
	var d = time.Since(startedAt)

	if err != nil {
		router.log.Infov(1, "error routing %s query %v -> %v:%v after %v: %v",
			query.Origin(), query.Caller(), query.Target(), query.Query(), d, err,
		)
		if rnf, ok := err.(*net.ErrRouteNotFound); ok {
			for _, line := range strings.Split(rnf.Trace(), "\n") {
				if len(line) > 0 {
					router.log.Logv(2, "%v", line)
				}
			}
		}
	} else {
		router.log.Infov(1, "routed %s query %v -> %v:%v in %v",
			query.Origin(), query.Caller(), target.Identity(), query.Query(), d,
		)
	}

	return target, err
}

func (router *CoreRouter) Conns() *ConnSet {
	return router.Monitor.Conns()
}
