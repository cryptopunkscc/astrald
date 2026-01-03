package apphost

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (mod *Module) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	for _, handler := range mod.handlers.Clone() {
		if !handler.Identity.IsEqual(q.Target) {
			continue
		}

		conn, err := handler.RouteQuery(ctx, q, w)

		// check response
		var rejected *astral.ErrRejected
		switch {
		case err == nil: // accepted
			return conn, nil

		case errors.As(err, &rejected): // rejected
			return query.RejectWithCode(rejected.Code)

		case errors.Is(err, errEndpointUnavailable):
			mod.log.Logv(3, "removing unresponsive query handler at %v", handler.Endpoint)
			mod.handlers.Remove(handler)
		}
	}

	return query.RouteNotFound(mod)
}
