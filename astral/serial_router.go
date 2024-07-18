package astral

import (
	"context"
	"errors"
)

// SerialRouter tries to route queries through every router in the array and returns first successful attempt or
// ErrRouteNotFound. Use Trace() method of the errors to see details of the failed routing attempt.
type SerialRouter struct {
	Routers []Router
}

func NewSerialRouter(routers ...Router) *SerialRouter {
	return &SerialRouter{Routers: routers}
}

func (m *SerialRouter) RouteQuery(ctx context.Context, query Query, caller SecureWriteCloser, hints Hints) (SecureWriteCloser, error) {
	var rerr = &ErrRouteNotFound{Router: m}
	for _, router := range m.Routers {
		target, err := router.RouteQuery(ctx, query, caller, hints)
		if err == nil {
			return target, nil
		}
		if errors.Is(err, ErrRejected) {
			return nil, err
		}
		rerr.Fails = append(rerr.Fails, err)
	}
	return nil, rerr
}

func (m *SerialRouter) AddRouter(router Router) {
	m.Routers = append(m.Routers, router)
}
