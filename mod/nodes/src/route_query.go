package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// see if we already have a link with the target
	for _, link := range mod.links.Clone() {
		if link.RemoteIdentity().IsEqual(query.Target()) {
			return link.RouteQuery(ctx, query, caller, hints)
		}
	}

	// if not, try to establish a new link with the target
	link, err := mod.Link(ctx, query.Target(), nodes.LinkOpts{})
	if err == nil {
		return link.RouteQuery(ctx, query, caller, hints)
	}

	mod.log.Errorv(2, "error linking with %v: %v", query.Target(), err)

	return net.RouteNotFound(mod, err)
}
