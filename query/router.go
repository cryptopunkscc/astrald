package query

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"io"
)

type Router interface {
	RouteQuery(ctx context.Context, q Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error)
}

func Accept(query Query, remoteWriter net.SecureWriteCloser, handler func(conn net.SecureConn)) (net.SecureWriteCloser, error) {
	r, wc := io.Pipe()

	go handler(net.NewSecureConn(remoteWriter, r, query.Caller()))

	return net.NewSecureWriteCloser(wc, query.Target()), nil
}

func Run(ctx context.Context, router Router, query Query) (net.SecureConn, error) {
	r, wc := io.Pipe()

	lswc := net.NewSecureWriteCloser(wc, query.Caller())

	rswc, err := router.RouteQuery(ctx, query, lswc)
	if err != nil {
		return nil, err
	}

	if !query.Target().IsEqual(rswc.RemoteIdentity()) {
		rswc.Close()
		return nil, errors.New("response identity mismatch")
	}

	return net.NewSecureConn(rswc, r, query.Caller()), err
}
