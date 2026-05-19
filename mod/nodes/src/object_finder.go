package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) FindObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *astral.Identity, error) {
	out := make(chan *astral.Identity, 1)
	defer close(out)

	if id, found := mod.searchCache.Get(id.String()); found {
		out <- id
	}

	return out, nil
}
