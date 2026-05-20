package apphost

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (mod *Module) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
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

	for _, h := range mod.wsHandlers.Clone() {
		if !h.Identity.IsEqual(q.Target) {
			continue
		}

		conn, err := h.RouteQuery(ctx, q, w)

		var rejected *astral.ErrRejected
		switch {
		case err == nil:
			return conn, nil
		case errors.As(err, &rejected):
			return query.RejectWithCode(rejected.Code)
		case errors.Is(err, errHandlerGone):
			mod.log.Logv(3, "removing closed ws handler for %v", h.Identity)
			mod.wsHandlers.Remove(h)
		}
	}

	return query.RouteNotFound()
}

func (mod *Module) removeHandlersByToken(token astral.Nonce) error {
	for _, h := range mod.handlers.Clone() {
		if h.IpcToken == token {
			mod.handlers.Remove(h)
		}
	}
	return nil
}
