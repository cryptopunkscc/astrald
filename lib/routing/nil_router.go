package routing

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

var _ astral.Router = &NilRouter{}

type NilRouter struct {
	Soft bool // return ErrRouteNotFound instead of ErrRejected
}

func (r *NilRouter) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	if r.Soft {
		return query.RouteNotFound(r, errors.New("nil router"))
	}
	return query.Reject()
}
