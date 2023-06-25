package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type PeerRouter struct {
	Target id.Identity
	*CoreNetwork
}

func NewPeerRouter(network *CoreNetwork, target id.Identity) *PeerRouter {
	return &PeerRouter{Target: target, CoreNetwork: network}
}

func (router *PeerRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	var links = router.links.ByRemoteIdentity(router.Target).ByLocalIdentity(query.Caller())
	var best = net.SelectLink(links.All(), BestQuality)

	if best == nil {
		best, _ = router.Link(ctx, query.Target())
	}

	if best == nil {
		return nil, &net.ErrRouteNotFound{Router: router}
	}

	return best.RouteQuery(ctx, query, caller)
}
