package routers

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Router = &routerFunc{}

type routerFunc struct {
	fn astral.RouteQueryFunc
}

func (r *routerFunc) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	return r.fn(ctx, q, w)
}

func Func(fn astral.RouteQueryFunc) astral.Router {
	return &routerFunc{fn: fn}
}
