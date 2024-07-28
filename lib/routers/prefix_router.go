package routers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
	"sync"
)

var _ LocalRouter = &PrefixRouter{}

// PrefixRouter is a LocalRouter that allows adding exact and prefix routes. Exact matches take priority, and when
// multiple prefix routes match, the longest prefix wins.
//
// Example:
//
// router.AddRoute("my_service", router)    // will only match the string "my_service"
// router.AddRoute("my_service.*", router)  // will match all queries that begin with "my_service."
type PrefixRouter struct {
	Exclusive    bool
	EnableParams bool
	exact        map[string]astral.Router
	prefix       map[string]astral.Router
	mu           sync.RWMutex
}

// NewPrefixRouter makes a new PrefixRouter. If exclusive is true, it will reject unmatched queries, otherwise it
// will return the RouteNotFound error.
func NewPrefixRouter(exclusive bool) *PrefixRouter {
	return &PrefixRouter{
		exact:     make(map[string]astral.Router),
		prefix:    make(map[string]astral.Router),
		Exclusive: exclusive,
	}
}

func (router *PrefixRouter) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser, hints astral.Hints) (io.WriteCloser, error) {
	var baseQuery = query.Query
	if router.EnableParams {
		if i := strings.IndexByte(baseQuery, '?'); i != -1 {
			baseQuery = baseQuery[:i]
		}
	}

	route := router.Match(baseQuery)

	if route == nil {
		if router.Exclusive == true {
			return astral.Reject()
		} else {
			return astral.RouteNotFound(router)
		}
	}

	return route.RouteQuery(ctx, query, caller, hints)
}

// AddRoute adds a route to the router
func (router *PrefixRouter) AddRoute(name string, target astral.Router) error {
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
func (router *PrefixRouter) AddRouteFunc(name string, fn astral.RouteQueryFunc) error {
	return router.AddRoute(name, Func(fn))
}

// RemoveRoute removes a route from the router
func (router *PrefixRouter) RemoveRoute(name string) error {
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
func (router *PrefixRouter) Routes() []PathRoute {
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
func (router *PrefixRouter) Match(query string) astral.Router {
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

// LocalRouter is a router that routes queries for a single local identity
type LocalRouter interface {
	astral.Router
	AddRoute(name string, target astral.Router) error
	RemoveRoute(name string) error
	Routes() []PathRoute
	Match(query string) astral.Router
}
