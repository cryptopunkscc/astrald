package core

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"strings"
	"sync"
)

var _ node.LocalRouter = &PrefixRouter{}

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
	exact        map[string]net.Router
	prefix       map[string]net.Router
	mu           sync.RWMutex
}

// NewPrefixRouter makes a new PrefixRouter. If exclusive is true, it will reject unmatched queries, otherwise it
// will return the RouteNotFound error.
func NewPrefixRouter(exclusive bool) *PrefixRouter {
	return &PrefixRouter{
		exact:     make(map[string]net.Router),
		prefix:    make(map[string]net.Router),
		Exclusive: exclusive,
	}
}

func (router *PrefixRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var baseQuery = query.Query()
	if router.EnableParams {
		if i := strings.IndexByte(baseQuery, '?'); i != -1 {
			baseQuery = baseQuery[:i]
		}
	}

	route := router.Match(baseQuery)

	if route == nil {
		if router.Exclusive == true {
			return net.Reject()
		} else {
			return net.RouteNotFound(router)
		}
	}

	return route.RouteQuery(ctx, query, caller, hints)
}

// AddRoute adds a route to the router
func (router *PrefixRouter) AddRoute(name string, target net.Router) error {
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
func (router *PrefixRouter) AddRouteFunc(name string, fn net.RouteQueryFunc) error {
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
func (router *PrefixRouter) Routes() []node.LocalRoute {
	var routes []node.LocalRoute

	for k, v := range router.exact {
		routes = append(routes, node.LocalRoute{
			Name:   k,
			Target: v,
		})
	}

	for k, v := range router.prefix {
		routes = append(routes, node.LocalRoute{
			Name:   k + "*",
			Target: v,
		})
	}

	return routes
}

// Match finds the best match for the query and returns the router
func (router *PrefixRouter) Match(query string) net.Router {
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
