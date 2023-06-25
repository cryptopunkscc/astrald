package net

import (
	"context"
	"errors"
	"io"
)

type Router interface {
	RouteQuery(ctx context.Context, query Query, caller SecureWriteCloser) (SecureWriteCloser, error)
}

// Accept accepts the query and runs the handler in a new goroutine.
func Accept(query Query, caller SecureWriteCloser, handler func(conn SecureConn)) (SecureWriteCloser, error) {
	r, wc := io.Pipe()

	go handler(NewSecureConn(caller, r, query.Target()))

	return NewSecureWriteCloser(wc, query.Target()), nil
}

// Route routes a query through the provided Router. It returns a SecureConn if query was successfully routed
// to the target and accepted, otherwise it returns an error.
// Errors: ErrRouteNotFound ErrRejected ...
func Route(ctx context.Context, router Router, query Query) (SecureConn, error) {
	localReader, wc := io.Pipe()

	remoteWriter := NewSecureWriteCloser(wc, query.Caller())

	localWriter, err := router.RouteQuery(ctx, query, remoteWriter)
	if err != nil {
		return nil, err
	}

	if !query.Target().IsZero() && !query.Target().IsEqual(localWriter.Identity()) {
		localWriter.Close()
		return nil, errors.New("response identity mismatch")
	}

	return NewSecureConn(localWriter, localReader, query.Caller()), err
}
