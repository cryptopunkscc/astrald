package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) FindObject(ctx *astral.Context, id *object.ID, scope *astral.Scope) (sources []*astral.Identity) {
	if id, found := mod.searchCache.Get(id.String()); found {
		sources = append(sources, id)
	}
	return
}
