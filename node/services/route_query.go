package services

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
)

func (srv *CoreServices) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// Fetch the service
	serviceName := query.Query()
	if idx := strings.IndexByte(serviceName, '?'); idx != -1 {
		serviceName = serviceName[0:idx]
	}

	service, err := srv.Find(query.Target(), serviceName)
	if err != nil {
		return net.RouteNotFound(srv)
	}

	if service.Router == nil {
		return nil, errors.New("service unreachable")
	}

	target, err := service.RouteQuery(ctx, query, caller, hints)
	if err != nil {
		return nil, err
	}

	if !target.Identity().IsEqual(query.Target()) && !hints.AllowRedirect {
		target.Close()
		return net.RouteNotFound(srv, errors.New("response identity mismatch"))
	}

	return target, err
}
