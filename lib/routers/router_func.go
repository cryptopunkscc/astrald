package routers

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Router = &routerFunc{}

type routerFunc struct {
	fn astral.RouteQueryFunc
}

func (r *routerFunc) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	return r.fn(ctx, query, caller)
}

func Func(fn astral.RouteQueryFunc) astral.Router {
	return &routerFunc{fn: fn}
}
