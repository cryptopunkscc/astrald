package routers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
	"sync"
)

type PathRouter struct {
	identity  *astral.Identity
	Authority bool
	exact     map[string]astral.Router
	prefix    map[string]astral.Router
	mu        sync.RWMutex
}

// NewPathRouter makes a new PathRouter.
// If identity is not zero only queries to this identity will be routed, other queries will end with ErrRouteNotFound.
// If authority is true, ErrRejected will be returned in case of routing failure (instead of ErrRouteNotFound).
func NewPathRouter(identity *astral.Identity, authority bool) *PathRouter {
	return &PathRouter{
		identity:  identity,
		exact:     make(map[string]astral.Router),
		prefix:    make(map[string]astral.Router),
		Authority: authority,
	}
}

func (router *PathRouter) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if !router.identity.IsZero() {
		if !query.Target.IsEqual(router.identity) {
			return astral.RouteNotFound(router)
		}
	}

	var baseQuery = query.Query

	if i := strings.IndexByte(baseQuery, '?'); i != -1 {
		baseQuery = baseQuery[:i]
	}

	route := router.Match(baseQuery)

	if route == nil {
		if router.Authority == true {
			return astral.Reject()
		} else {
			return astral.RouteNotFound(router)
		}
	}

	return route.RouteQuery(ctx, query, caller)
}

// AddRoute adds a route to the router
func (router *PathRouter) AddRoute(name string, target astral.Router) error {
	router.mu.Lock()
	defer router.mu.Unlock()

	if len(name) == 0 {
		return errors.New("invalid name")
	}

	if name[len(name)-1] == '*' {
		name = name[:len(name)-1]

		if _, found := router.prefix[name]; found {
			return errors.New("prefix already added")
		}

		router.prefix[name] = target
	} else {
		if _, found := router.exact[name]; found {
			return errors.New("name already added")
		}

		router.exact[name] = target
	}

	return nil
}

// AddRouteFunc adds a route handler function to the router
func (router *PathRouter) AddRouteFunc(name string, fn astral.RouteQueryFunc) error {
	return router.AddRoute(name, Func(fn))
}

// RemoveRoute removes a route from the router
func (router *PathRouter) RemoveRoute(name string) error {
	router.mu.Lock()
	defer router.mu.Unlock()

	if len(name) == 0 {
		return errors.New("invalid name")
	}

	if name[len(name)-1] == '*' {
		name = name[:len(name)-1]

		if _, found := router.prefix[name]; found {
			delete(router.prefix, name)
			return nil
		}
	} else {
		if _, found := router.exact[name]; found {
			delete(router.exact, name)
			return nil
		}
	}

	return errors.New("route not found")
}

// Routes returns an unordered list of all routes
func (router *PathRouter) Routes() []PathRoute {
	var routes []PathRoute

	for k, v := range router.exact {
		routes = append(routes, PathRoute{
			Name:   k,
			Target: v,
		})
	}

	for k, v := range router.prefix {
		routes = append(routes, PathRoute{
			Name:   k + "*",
			Target: v,
		})
	}

	return routes
}

// Match finds the best match for the query and returns the router
func (router *PathRouter) Match(query string) astral.Router {
	router.mu.RLock()
	defer router.mu.RUnlock()

	// check exact matches first
	if route, found := router.exact[query]; found {
		return route
	}

	// find the best (longest) matching prefix route
	var best astral.Router
	var bestLen int
	for prefix, route := range router.prefix {
		if strings.HasPrefix(query, prefix) {
			if best == nil || len(prefix) > bestLen {
				best = route
				bestLen = len(prefix)
			}
		}
	}

	return best
}

type PathRoute struct {
	Name   string
	Target astral.Router
}
