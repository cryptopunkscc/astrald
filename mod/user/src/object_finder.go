package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) FindObject(ctx *astral.Context, id *object.ID, scope *astral.Scope) []*astral.Identity {
	return mod.listSibs()
}
