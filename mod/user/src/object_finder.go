package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) FindObject(ctx *astral.Context, id *astral.ObjectID, scope *astral.Scope) []*astral.Identity {
	return mod.getLinkedSibs()
}
