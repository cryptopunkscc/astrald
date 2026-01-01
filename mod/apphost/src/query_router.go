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
		case err == nil:
			return conn, nil
		case errors.As(err, &rejected):
			return query.RejectWithCode(rejected.Code)
		}
	}

	return query.RouteNotFound(mod)
}
