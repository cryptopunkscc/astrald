package core

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.Router = &routerFunc{}

type routerFunc struct {
	fn net.RouteQueryFunc
}

func (r *routerFunc) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return r.fn(ctx, query, caller, hints)
}

func Func(fn net.RouteQueryFunc) net.Router {
	return &routerFunc{fn: fn}
}
