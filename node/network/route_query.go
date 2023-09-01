package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (n *CoreNetwork) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (targetWriter net.SecureWriteCloser, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if query.Caller().IsZero() {
		return nil, errors.New("caller has zero value")
	}

	return NewPeerRouter(n, query.Target()).RouteQuery(ctx, query, caller, hints)
}
