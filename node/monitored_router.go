package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

type MonitoredRouter struct {
	net.Router
	conns *ConnSet
}

func (router *MonitoredRouter) Conns() *ConnSet {
	return router.conns
}

func NewMonitoredRouter(router net.Router) *MonitoredRouter {
	return &MonitoredRouter{
		Router: router,
		conns:  NewConnSet(),
	}
}

func (router *MonitoredRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
	var callerMonitor = NewMonitoredWriter(caller)

	target, err = router.Router.RouteQuery(ctx, query, callerMonitor, hints)
	if err != nil {
		return nil, err
	}

	var targetMonitor = NewMonitoredWriter(target)
	var conn = NewConn(callerMonitor, targetMonitor, query, hints)

	router.conns.Add(conn)
	go func() {
		<-conn.Done()
		router.conns.Remove(conn)
	}()

	return targetMonitor, err
}
