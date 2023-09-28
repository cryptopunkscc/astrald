package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"time"
)

type CoreRouter struct {
	Routers       *net.SerialRouter
	Monitor       *MonitoredRouter
	log           *log.Logger
	logRouteTrace bool
}

func NewCoreRouter(routers []net.Router, log *log.Logger) *CoreRouter {
	var router = &CoreRouter{
		Routers: net.NewSerialRouter(routers...),
		log:     log,
	}

	router.Monitor = NewMonitoredRouter(router.Routers)

	return router
}

func (router *CoreRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	router.log.Logv(
		2,
		"[%v] %v -> %v:%v routing...",
		query.Nonce(),
		query.Caller(),
		query.Target(),
		query.Query(),
	)

	var startedAt = time.Now()
	target, err := router.Monitor.RouteQuery(ctx, query, caller, hints)
	var d = time.Since(startedAt).Round(1 * time.Microsecond)

	if err != nil {
		router.log.Infov(
			0,
			"[%v] %v -> %v:%v error (%v): %v",
			query.Nonce(),
			query.Caller(),
			query.Target(),
			query.Query(),
			d,
			err,
		)

		if router.logRouteTrace {
			if rnf, ok := err.(*net.ErrRouteNotFound); ok {
				for _, line := range strings.Split(rnf.Trace(), "\n") {
					if len(line) > 0 {
						router.log.Logv(2, "%v", line)
					}
				}
			}
		}
	} else {
		router.log.Infov(
			0,
			"[%v] %v -> %v:%v routed in %v",
			query.Nonce(),
			query.Caller(),
			target.Identity(),
			query.Query(),
			d,
		)
	}

	return target, err
}

func (router *CoreRouter) Conns() *ConnSet {
	return router.Monitor.Conns()
}

func (router *CoreRouter) LogRouteTrace() bool {
	return router.logRouteTrace
}

func (router *CoreRouter) SetLogRouteTrace(logRouteTrace bool) {
	router.logRouteTrace = logRouteTrace
}
