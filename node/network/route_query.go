package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

func (n *CoreNetwork) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (targetWriter net.SecureWriteCloser, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if query.Target().IsEqual(n.node.Identity()) {
		return net.RouteNotFound(n)
	}

	if query.Caller().IsZero() {
		return net.RouteNotFound(n, errors.New("caller identity missing in query"))
	}

	var links = n.links.ByRemoteIdentity(query.Target()).All()
	if len(links) > 0 {
		return links[0].RouteQuery(ctx, query, caller, hints)
	}

	lnk, err := link.MakeLink(ctx, n.node, query.Target(), link.Opts{})
	if err != nil {
		return net.RouteNotFound(n)
	}

	err = n.AddLink(lnk)
	if err != nil {
		return net.RouteNotFound(n)
	}

	return lnk.RouteQuery(ctx, query, caller, hints)
}
