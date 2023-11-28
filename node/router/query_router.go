package router

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
	"sync"
)

type QueryRouter struct {
	exact  map[string]net.Router
	prefix map[string]net.Router
	mu     sync.RWMutex
}

type QueryRoute struct {
	Name   string
	Router net.Router
}

func NewQueryRouter() *QueryRouter {
	return &QueryRouter{
		exact:  make(map[string]net.Router),
		prefix: make(map[string]net.Router),
	}
}

func (router *QueryRouter) Match(query string) net.Router {
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

func (router *QueryRouter) AddRoute(name string, target net.Router) error {
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

func (router *QueryRouter) RemoveRoute(name string) error {
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

func (router *QueryRouter) Routes() []QueryRoute {
	var routes []QueryRoute

	for k, v := range router.exact {
		routes = append(routes, QueryRoute{
			Name:   k,
			Router: v,
		})
	}

	for k, v := range router.prefix {
		routes = append(routes, QueryRoute{
			Name:   k + "*",
			Router: v,
		})
	}

	return routes
}

func (router *QueryRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	route := router.Match(query.Query())

	if route == nil {
		return net.Reject()
	}

	return route.RouteQuery(ctx, query, caller, hints)
}
