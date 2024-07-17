package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

type LinkedFinder struct {
	mod *Module
}

func (finder *LinkedFinder) Find(ctx context.Context, objectID object.ID, scope *net.Scope) (sources []id.Identity) {
	return finder.mod.nodes.Peers()
}
