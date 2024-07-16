package net

import (
	"context"
	"errors"
)

type Router interface {
	RouteQuery(ctx context.Context, query Query, caller SecureWriteCloser, hints Hints) (SecureWriteCloser, error)
}

type RouteQueryFunc func(context.Context, Query, SecureWriteCloser, Hints) (SecureWriteCloser, error)

var _ Router = NilRouter{}

type NilRouter struct {
	Soft bool // return ErrRouteNotFound instead of ErrRejected
}

func (r NilRouter) RouteQuery(ctx context.Context, query Query, caller SecureWriteCloser, hints Hints) (SecureWriteCloser, error) {
	if r.Soft {
		return RouteNotFound(r, errors.New("nil router"))
	}
	return Reject()
}

// Accept accepts the query and runs the handler in a new goroutine.
func Accept(query Query, src SecureWriteCloser, handler func(conn Conn)) (SecureWriteCloser, error) {
	pipeReader, pipeWriter := SecurePipe(query.Target())

	go handler(NewSecureConn(src, pipeReader, false))

	return pipeWriter, nil
}

func Reject() (SecureWriteCloser, error) {
	return nil, ErrRejected
}

func Abort() (SecureWriteCloser, error) {
	return nil, ErrAborted
}

// Route routes a query through the provided Router. It returns a SecureConn if query was successfully routed
// to the target and accepted, otherwise it returns an error.
// Errors: ErrRouteNotFound ErrRejected ...
func Route(ctx context.Context, router Router, query Query) (Conn, error) {
	return RouteWithHints(ctx, router, query, DefaultHints())
}

func RouteWithHints(ctx context.Context, router Router, query Query, hints Hints) (Conn, error) {
	pipeReader, pipeWriter := SecurePipe(query.Caller())

	target, err := router.RouteQuery(ctx, query, pipeWriter, hints)
	if err != nil {
		return nil, err
	}

	if !query.Target().IsEqual(target.Identity()) {
		target.Close()
		return nil, errors.New("response identity mismatch")
	}

	return NewSecureConn(target, pipeReader, true), err
}
