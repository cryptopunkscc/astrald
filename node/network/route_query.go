package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

func (n *CoreNetwork) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (targetWriter net.SecureWriteCloser, err error) {
	// don't open a new link with the target if we have one already
	if n.links.ByRemoteIdentity(query.Target()).Count() > 0 {
		return net.RouteNotFound(n)
	}

	lnk, err := link.MakeLink(ctx, n.node, query.Target(), link.Opts{})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, net.ErrAborted

		case errors.Is(err, context.DeadlineExceeded):
			return nil, net.ErrTimeout
		}

		return net.RouteNotFound(n, err)
	}

	err = n.AddLink(lnk)
	if err != nil {
		return net.RouteNotFound(n, err)
	}

	// reroute the query
	return n.node.Router().RouteQuery(ctx, query, caller, hints.SetReroute())
}
