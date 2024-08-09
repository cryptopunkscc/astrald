package astral

import (
	"context"
	"io"
)

type Router interface {
	RouteQuery(ctx context.Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
}

type RouteQueryFunc func(ctx context.Context, q *Query, w io.WriteCloser) (io.WriteCloser, error)
