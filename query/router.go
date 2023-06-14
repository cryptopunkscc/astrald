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
	localReader, wc := io.Pipe()

	remoteWriter := net.NewSecureWriteCloser(wc, query.Caller())

	localWriter, err := router.RouteQuery(ctx, query, remoteWriter)
	if err != nil {
		return nil, err
	}

	if !query.Target().IsEqual(localWriter.RemoteIdentity()) {
		localWriter.Close()
		return nil, errors.New("response identity mismatch")
	}

	return net.NewSecureConn(localWriter, localReader, query.Caller()), err
}
