package astral

import (
	"context"
	"errors"
	"io"
)

type Router interface {
	RouteQuery(ctx context.Context, query *Query, caller io.WriteCloser) (io.WriteCloser, error)
}

type RouteQueryFunc func(context.Context, *Query, io.WriteCloser) (io.WriteCloser, error)

var _ Router = NilRouter{}

type NilRouter struct {
	Soft bool // return ErrRouteNotFound instead of ErrRejected
}

func (r NilRouter) RouteQuery(ctx context.Context, query *Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if r.Soft {
		return RouteNotFound(r, errors.New("nil router"))
	}
	return Reject()
}

// Accept accepts the query and runs the handler in a new goroutine.
func Accept(query *Query, src io.WriteCloser, handler func(Conn)) (io.WriteCloser, error) {
	pipeReader, pipeWriter := io.Pipe()

	go handler(newConn(query.Target, query.Caller, src, pipeReader, false))

	return pipeWriter, nil
}

func Reject() (io.WriteCloser, error) {
	return nil, ErrRejected
}

func Abort() (io.WriteCloser, error) {
	return nil, ErrAborted
}

// Route routes a query through the provided Router. It returns a SecureConn if query was successfully routed
// to the target and accepted, otherwise it returns an error.
// Errors: ErrRouteNotFound ErrRejected ...
func Route(ctx context.Context, router Router, query *Query) (Conn, error) {
	pipeReader, pipeWriter := io.Pipe()

	target, err := router.RouteQuery(ctx, query, pipeWriter)
	if err != nil {
		return nil, err
	}

	return newConn(query.Caller, query.Target, target, pipeReader, true), err
}
