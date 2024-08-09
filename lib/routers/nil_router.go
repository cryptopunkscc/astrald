package routers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

var _ astral.Router = &NilRouter{}

type NilRouter struct {
	Soft bool // return ErrRouteNotFound instead of ErrRejected
}

func (r *NilRouter) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if r.Soft {
		return astral.RouteNotFound(r, errors.New("nil router"))
	}
	return query.Reject()
}
