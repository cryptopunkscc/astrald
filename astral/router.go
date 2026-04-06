package astral

import (
	"io"
)

type Router interface {
	RouteQuery(ctx *Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
}

type HasRoutingPriority interface {
	RoutingPriority() int
}

const (
	RoutingPriorityHigh   = 0
	RoutingPriorityMedium = 1000
	RoutingPriorityNormal = 2000
	RoutingPriorityLow    = 3000
)

type RouteQueryFunc func(ctx *Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
