package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
)

func (node *CoreNode) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return node.routes.RouteQuery(ctx, query, caller, hints)
}

func (node *CoreNode) AddRoute(name string, target net.Router) error {
	return node.routes.AddRoute(name, target)
}

func (node *CoreNode) RemoveRoute(name string) error {
	return node.routes.RemoveRoute(name)
}

func (node *CoreNode) Routes() []router.LocalRoute {
	return node.routes.Routes()
}
