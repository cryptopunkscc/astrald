package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

func (node *CoreNode) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return node.routes.RouteQuery(ctx, query, caller, hints)
}
