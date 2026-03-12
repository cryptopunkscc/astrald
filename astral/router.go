package astral

import (
	"io"
)

type Router interface {
	RouteQuery(ctx *Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
}

type RouteQueryFunc func(ctx *Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
