package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type MonitoredRouter struct {
	net.Router
	log   *log.Logger
	conns *ConnSet
}

func (router *MonitoredRouter) Conns() *ConnSet {
	return router.conns
}

func NewMonitoredRouter(router net.Router, log *log.Logger) *MonitoredRouter {
	return &MonitoredRouter{
		Router: router,
		conns:  NewConnSet(),
		log:    log,
	}
}

func (router *MonitoredRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (target net.SecureWriteCloser, err error) {
	router.log.Logv(2, "routing query %v -> %v:%v (origin %s)...", query.Caller(), query.Target(), query.Query(), query.Origin())

	var mcaller = NewMonitoredWriter(caller)

	var startedAt = time.Now()
	target, err = router.Router.RouteQuery(ctx, query, mcaller)
	var d = time.Since(startedAt)

	if err != nil {
		router.log.Infov(1, "error routing %s query %v -> %v:%v after %v: %v",
			query.Origin(), query.Caller(), query.Target(), query.Query(), d, err,
		)
		return nil, err
	}

	router.log.Infov(1, "routed %s query %v -> %v:%v in %v",
		query.Origin(), query.Caller(), target.Identity(), query.Query(), d,
	)

	var mtarget = NewMonitoredWriter(target)

	var conn = NewConn(mcaller, mtarget, query)

	router.conns.Add(conn)
	go func() {
		<-conn.Done()
		router.conns.Remove(conn)
	}()

	return mtarget, err
}
