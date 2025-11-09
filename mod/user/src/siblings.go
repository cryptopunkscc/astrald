package user

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Sibling struct {
	Id     *astral.Identity
	Cancel context.CancelFunc
}

func (mod *Module) notifyLinkedSibs(event string) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	for _, sib := range mod.getLinkedSibs() {
		sib := sib
		go mod.Objects.Push(mod.ctx, sib, &user.Notification{Event: astral.String8(event)})
	}
}

func (mod *Module) pushToLinkedSibs(object astral.Object) {
	ac := mod.ActiveContract()
	if ac == nil {
		return
	}

	for _, sib := range mod.getLinkedSibs() {
		sib := sib
		go mod.Objects.Push(mod.ctx, sib, object)
	}
}

func (mod *Module) getLinkedSibs() (list []*astral.Identity) {
	for _, sib := range mod.sibs.Values() {
		list = append(list, sib.Id)
	}

	return list
}

func (mod *Module) addSibling(id *astral.Identity,
	cancel context.CancelFunc) {
	mod.sibs.Set(id.String(), Sibling{
		Id:     id,
		Cancel: cancel,
	})
}
