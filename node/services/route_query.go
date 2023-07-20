package services

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (srv *CoreServices) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	// Fetch the service
	service, err := srv.Find(query.Target(), query.Query())
	if err != nil {
		return nil, &net.ErrRouteNotFound{Router: srv}
	}

	if service.Router == nil {
		return nil, errors.New("service unreachable")
	}

	target, err := service.RouteQuery(ctx, query, caller)
	if err != nil {
		return nil, err
	}

	if !target.Identity().IsEqual(query.Target()) {
		target.Close()
		return nil, errors.New("response identity mismatch")
	}

	return target, err
}
