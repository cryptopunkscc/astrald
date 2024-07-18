package routers

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Router = &routerFunc{}

type routerFunc struct {
	fn astral.RouteQueryFunc
}

func (r *routerFunc) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return r.fn(ctx, query, caller, hints)
}

func Func(fn astral.RouteQueryFunc) astral.Router {
	return &routerFunc{fn: fn}
}
