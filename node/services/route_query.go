package services

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
)

func (srv *CoreServices) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	// Fetch the service
	service, err := srv.Find(q.Query())
	if err != nil {
		return nil, err
	}

	if !q.Target().IsZero() && !q.Target().IsEqual(service.identity) {
		return nil, errors.New("target identity mismatch")
	}

	if q.Target().IsZero() {
		q = query.NewOrigin(q.Caller(), service.identity, q.Query(), q.Origin())
	}

	if service.router == nil {
		return nil, errors.New("service unreachable")
	}

	localWriter, err := service.router.RouteQuery(ctx, q, remoteWriter)
	if err != nil {
		return nil, err
	}

	if !localWriter.RemoteIdentity().IsEqual(q.Target()) {
		localWriter.Close()
		return nil, errors.New("response identity mismatch")
	}

	return localWriter, err
}
