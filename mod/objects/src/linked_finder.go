package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/object"
)

type LinkedFinder struct {
	mod *Module
}

func (finder *LinkedFinder) Find(ctx context.Context, objectID object.ID, scope *astral.Scope) (sources []id.Identity) {
	return finder.mod.Nodes.Peers()
}
