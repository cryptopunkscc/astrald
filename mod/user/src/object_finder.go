package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) FindObject(ctx *astral.Context, id *astral.ObjectID) []*astral.Identity {
	return mod.getLinkedSibs()
}
