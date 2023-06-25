package services

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (srv *CoreServices) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	// Fetch the service
	service, err := srv.Find(query.Query())
	if err != nil {
		return nil, err
	}

	if !query.Target().IsZero() && !query.Target().IsEqual(service.identity) {
		return nil, errors.New("target identity mismatch")
	}

	if query.Target().IsZero() {
		query = net.NewOrigin(query.Caller(), service.identity, query.Query(), query.Origin())
	}

	if service.router == nil {
		return nil, errors.New("service unreachable")
	}

	target, err := service.router.RouteQuery(ctx, query, caller)
	if err != nil {
		return nil, err
	}

	if !target.Identity().IsEqual(query.Target()) {
		target.Close()
		return nil, errors.New("response identity mismatch")
	}

	return target, err
}
