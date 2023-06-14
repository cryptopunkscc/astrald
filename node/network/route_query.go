package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
)

func (n *CoreNetwork) RouteQuery(ctx context.Context, query query.Query, w net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if peer := n.Peers().Find(query.Target()); peer != nil {
		peer.Check()
	}

	l, err := n.Link(ctx, query.Target())
	if err != nil {
		return nil, err
	}

	return l.RouteQuery(ctx, query, w)
}
