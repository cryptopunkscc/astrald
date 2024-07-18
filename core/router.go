package core

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/astral"
	"strings"
	"sync"
	"time"
)

var _ astral.Router = &Router{}

const routingTimeout = time.Minute
const ViaRouterHintKey = "via"

type Router struct {
	*routers.PriorityRouter
	log           *log.Logger
	events        events.Queue
	conns         *ConnSet
	mu            sync.RWMutex
	logRouteTrace bool
	enroute       map[string]struct{}
	enrouteMu     sync.Mutex
}

func NewRouter(log *log.Logger, eventParent *events.Queue) *Router {
	var router = &Router{
		PriorityRouter: routers.NewPriorityRouter(),
		conns:          NewConnSet(),
		enroute:        map[string]struct{}{},
		log:            log,
	}
	router.events.SetParent(eventParent)

	return router
}

func (r *Router) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (target astral.SecureWriteCloser, err error) {
	var silent = hints.Silent

	if !r.lockNonce(query.Nonce()) {
		if !hints.Reroute {
			return astral.RouteNotFound(r, errors.New("routing cycle not allowed"))
		}
		silent = true
	}
	defer r.unlockNonce(query.Nonce())

	// log the start of routing
	if !silent {
		r.log.Logv(2, "[%v] %v -> %v:%v routing...",
			query.Nonce(),
			query.Caller(),
			query.Target(),
			query.Query(),
		)
	}

	var startedAt = time.Now()

	if hints.Reroute {
		hints.Reroute = false
		var conn = r.conns.FindByNonce(query.Nonce())
		if conn == nil {
			return nil, errors.New("rerouted nonce does not exist")
		}

		// update the query in monitored connection
		if hints.Update {
			conn.query = query
		}

		target, err = r.routeQuery(ctx, query, caller, hints)
	} else {
		ctx, cancel := context.WithTimeout(ctx, routingTimeout)
		defer cancel()

		target, err = r.routeMonitored(ctx, query, caller, hints)
	}

	var d = time.Since(startedAt).Round(1 * time.Microsecond)

	// log routing results
	if !silent {
		if err != nil {
			r.log.Infov(0, "[%v] %v -> %v:%v error (%v): %v",
				query.Nonce(),
				query.Caller(),
				query.Target(),
				query.Query(),
				d,
				err,
			)

			if r.logRouteTrace {
				if rnf, ok := err.(*astral.ErrRouteNotFound); ok {
					for _, line := range strings.Split(rnf.Trace(), "\n") {
						if len(line) > 0 {
							r.log.Logv(2, "[%v] %v", query.Nonce(), line)
						}
					}
				}
			}
		} else {
			r.log.Infov(0, "[%v] %v -> %v:%v routed in %v",
				query.Nonce(),
				query.Caller(),
				target.Identity(),
				query.Query(),
				d,
			)
		}
	}

	return target, err
}

func (r *Router) routeMonitored(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (target astral.SecureWriteCloser, err error) {
	// monitor the caller
	var callerMonitor = NewMonitoredWriter(caller)

	// prepare the en route connection
	var conn = NewMonitoredConn(callerMonitor, nil, query, hints)
	r.conns.Add(conn)

	// route to next hop
	target, err = r.routeQuery(ctx, query, callerMonitor, hints)
	if err != nil {
		r.conns.Remove(conn)
		return nil, err
	}

	// monitor the target
	var targetMonitor = NewMonitoredWriter(target)
	conn.SetTarget(targetMonitor)

	r.events.Emit(EventConnAdded{Conn: conn})

	// remove the conn after it's closed
	go func() {
		<-conn.Done()
		r.conns.Remove(conn)
		r.events.Emit(EventConnRemoved{Conn: conn})
	}()

	return targetMonitor, err
}

func (r *Router) routeQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (target astral.SecureWriteCloser, err error) {
	if via, found := hints.Value(ViaRouterHintKey); found {
		switch typed := via.(type) {
		case astral.RouteQueryFunc:
			return typed(ctx, query, caller, hints)
		case astral.Router:
			return typed.RouteQuery(ctx, query, caller, hints)
		default:
			return astral.RouteNotFound(r, errors.New("via: unknown router type"))
		}
	}

	return r.PriorityRouter.RouteQuery(ctx, query, caller, hints)
}

func (r *Router) LogRouteTrace() bool {
	return r.logRouteTrace
}

func (r *Router) SetLogRouteTrace(logRouteTrace bool) {
	r.logRouteTrace = logRouteTrace
}

func (r *Router) lockNonce(nonce astral.Nonce) bool {
	r.enrouteMu.Lock()
	defer r.enrouteMu.Unlock()

	_, found := r.enroute[nonce.String()]
	if found {
		return false
	}
	r.enroute[nonce.String()] = struct{}{}
	return true
}

func (r *Router) unlockNonce(nonce astral.Nonce) bool {
	r.enrouteMu.Lock()
	defer r.enrouteMu.Unlock()

	_, found := r.enroute[nonce.String()]
	if !found {
		return false
	}
	delete(r.enroute, nonce.String())
	return true
}
