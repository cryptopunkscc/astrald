package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) FindObject(ctx *astral.Context, id *astral.ObjectID, scope *astral.Scope) (sources []*astral.Identity) {
	if id, found := mod.searchCache.Get(id.String()); found {
		sources = append(sources, id)
	}
	return
}
