package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
)

func (node *CoreNode) RouteQuery(ctx context.Context, query query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if query.Target().IsEqual(node.identity) {
		return node.Services().RouteQuery(ctx, query, remoteWriter)
	}
	return node.Network().RouteQuery(ctx, query, remoteWriter)
}
