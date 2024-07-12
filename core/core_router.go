package core

import (
	"cmp"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"slices"
	"strings"
	"sync"
	"time"
)

var _ node.Router = &CoreRouter{}

const routingTimeout = time.Minute

type CoreRouter struct {
	log           *log.Logger
	events        events.Queue
	conns         *ConnSet
	routes        []node.Route
	mu            sync.RWMutex
	logRouteTrace bool
	enroute       map[string]struct{}
	enrouteMu     sync.Mutex
}

func NewCoreRouter(log *log.Logger, eventParent *events.Queue) *CoreRouter {
	var router = &CoreRouter{
		conns:   NewConnSet(),
		routes:  make([]node.Route, 0),
		enroute: map[string]struct{}{},
		log:     log,
	}
	router.events.SetParent(eventParent)

	return router
}

func (r *CoreRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
	var silent = hints.Silent

	if !r.lockNonce(query.Nonce()) {
		if !hints.Reroute {
			return net.RouteNotFound(r, errors.New("routing cycle not allowed"))
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
				if rnf, ok := err.(*net.ErrRouteNotFound); ok {
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

func (r *CoreRouter) routeMonitored(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
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

func (r *CoreRouter) routeQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
	if via, found := hints.Value(ViaRouterHintKey); found {
		switch typed := via.(type) {
		case net.RouteQueryFunc:
			return typed(ctx, query, caller, hints)
		case net.Router:
			return typed.RouteQuery(ctx, query, caller, hints)
		default:
			return net.RouteNotFound(r, errors.New("via: unknown router type"))
		}
	}

	var routes = MatchRoutes(r.Routes(), query.Caller(), query.Target())
	var routingErrors []error

	for _, route := range routes {
		target, err = route.Router.RouteQuery(ctx, query, caller, hints)
		switch {
		case err == nil:
			return target, nil

		case errors.Is(err, net.ErrRejected),
			errors.Is(err, net.ErrAborted),
			errors.Is(err, net.ErrTimeout):
			return nil, err
		}

		routingErrors = append(routingErrors, err)
	}

	target, err = net.RouteNotFound(r, routingErrors...)

	return
}

func (r *CoreRouter) AddRoute(caller id.Identity, target id.Identity, router net.Router, priority int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// update route priority if route already exists...
	for _, route := range r.routes {
		if route.Router == router &&
			route.Caller.IsEqual(caller) &&
			route.Target.IsEqual(target) {
			route.Priority = priority
			return nil
		}
	}

	// ...if not, append a new route
	r.routes = append(r.routes, node.Route{
		Caller:   caller,
		Target:   target,
		Router:   router,
		Priority: priority,
	})

	return nil
}

func (r *CoreRouter) RemoveRoute(caller id.Identity, target id.Identity, router net.Router) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Router != router ||
			!route.Caller.IsEqual(caller) ||
			!route.Target.IsEqual(target) {
			continue
		}

		r.routes = append(r.routes[:i], r.routes[i+1:]...)
		return nil
	}

	return errors.New("route not found")
}

func (r *CoreRouter) Conns() *ConnSet {
	return r.conns
}

func (r *CoreRouter) Routes() []node.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes []node.Route
	for _, route := range r.routes {
		routes = append(routes, route)
	}

	slices.SortFunc(routes, func(a, b node.Route) int {
		return cmp.Compare(a.Priority, b.Priority) * -1
	})

	return routes
}

func (r *CoreRouter) LogRouteTrace() bool {
	return r.logRouteTrace
}

func (r *CoreRouter) SetLogRouteTrace(logRouteTrace bool) {
	r.logRouteTrace = logRouteTrace
}

func (r *CoreRouter) lockNonce(nonce net.Nonce) bool {
	r.enrouteMu.Lock()
	defer r.enrouteMu.Unlock()

	_, found := r.enroute[nonce.String()]
	if found {
		return false
	}
	r.enroute[nonce.String()] = struct{}{}
	return true
}

func (r *CoreRouter) unlockNonce(nonce net.Nonce) bool {
	r.enrouteMu.Lock()
	defer r.enrouteMu.Unlock()

	_, found := r.enroute[nonce.String()]
	if !found {
		return false
	}
	delete(r.enroute, nonce.String())
	return true
}
