package query

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

const DefaultRejectCode = 1

// Route routes a query through the provided Router. It returns a conn if query was successfully routed
// to the target and accepted, otherwise it returns an error.
// Errors: ErrRouteNotFound ErrRejected ...
func Route(ctx *astral.Context, r astral.Router, q *astral.Query) (astral.Conn, error) {
	ctx, cancel := ctx.WithTimeout(maxQueryTimeout)
	defer cancel()

	pipeReader, pipeWriter := io.Pipe()

	target, err := r.RouteQuery(ctx, q, pipeWriter)
	if err != nil {
		return nil, err
	}

	return newConn(q.Caller, q.Target, target, pipeReader, true), err
}

func RouteChan(ctx *astral.Context, r astral.Router, q *astral.Query) (*astral.Channel, error) {
	conn, err := Route(ctx, r, q)
	if err != nil {
		return nil, err
	}

	return astral.NewChannel(conn), nil
}

// Accept accepts the query and runs the handler in a new goroutine.
func Accept(query *astral.Query, src io.WriteCloser, handler func(astral.Conn)) (io.WriteCloser, error) {
	pipeReader, pipeWriter := io.Pipe()

	go handler(newConn(query.Target, query.Caller, src, pipeReader, false))

	return pipeWriter, nil
}

// Reject returns nil and an ErrRejected with the DefaultRejectCode.
func Reject() (io.WriteCloser, error) {
	return RejectWithCode(DefaultRejectCode)
}

// RejectWithCode returns nil and an ErrRejected with the given code.
// The code must not be 0.
func RejectWithCode(code uint8) (io.WriteCloser, error) {
	if code == 0 {
		panic("code cannot be 0")
	}
	return nil, &astral.ErrRejected{Code: code}
}

// RouteNotFound returns nil and an ErrRouteNotFound. r is the router that failed to route the query. errors are
// optional and will be wrapped.
func RouteNotFound(r astral.Router, errors ...error) (io.WriteCloser, error) {
	return nil, &astral.ErrRouteNotFound{
		Router: r,
		Fails:  errors,
	}
}
