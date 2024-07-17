package routers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"sync"
)

type PathRouter struct {
	identity  id.Identity
	Authority bool
	exact     map[string]net.Router
	prefix    map[string]net.Router
	mu        sync.RWMutex
}

// NewPathRouter makes a new PathRouter.
// If identity is not zero only queries to this identity will be routed, other queries will end with ErrRouteNotFound.
// If authority is true, ErrRejected will be returned in case of routing failure (instead of ErrRouteNotFound).
func NewPathRouter(identity id.Identity, authority bool) *PathRouter {
	return &PathRouter{
		identity:  identity,
		exact:     make(map[string]net.Router),
		prefix:    make(map[string]net.Router),
		Authority: authority,
	}
}

func (router *PathRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !router.identity.IsZero() {
		if query.Target().IsEqual(router.identity) {

		}
	}

	var baseQuery = query.Query()

	if i := strings.IndexByte(baseQuery, '?'); i != -1 {
		baseQuery = baseQuery[:i]
	}

	route := router.Match(baseQuery)

	if route == nil {
		if router.Authority == true {
			return net.Reject()
		} else {
			return net.RouteNotFound(router)
		}
	}

	return route.RouteQuery(ctx, query, caller, hints)
}

// AddRoute adds a route to the router
func (router *PathRouter) AddRoute(name string, target net.Router) error {
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
func (router *PathRouter) AddRouteFunc(name string, fn net.RouteQueryFunc) error {
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
func (router *PathRouter) Match(query string) net.Router {
	router.mu.RLock()
	defer router.mu.RUnlock()

	// check exact matches first
	if route, found := router.exact[query]; found {
		return route
	}

	// find the best (longest) matching prefix route
	var best net.Router
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
	Target net.Router
}
