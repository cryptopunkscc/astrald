package astral

import (
	"io"
)

// Router accepts a query and wires up the byte streams between caller and responder.
// w is the caller's response sink; the returned WriteCloser is the responder's request
// sink the caller writes into. A nil return with ErrRouteNotFound means the query was not handled.
type Router interface {
	RouteQuery(ctx *Context, q *InFlightQuery, w io.WriteCloser) (io.WriteCloser, error)
}

// HasRoutingPriority is the optional interface a Router implements to order itself among
// peers; lower values route first. Absence implies RoutingPriorityNormal.
type HasRoutingPriority interface {
	RoutingPriority() int
}

// Lower value routes first.
const (
	RoutingPriorityHigh   = 0
	RoutingPriorityMedium = 1000
	RoutingPriorityNormal = 2000
	RoutingPriorityLow    = 3000
)

type RouteQueryFunc func(ctx *Context, q *InFlightQuery, w io.WriteCloser) (io.WriteCloser, error)
