package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Sibling struct {
	Id *astral.Identity

	// NOTE: maybe could be scheduler.
	// Canellable if we have scheduler.Waitable
	MaintainLinkAction scheduler.ScheduledAction
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
	for _, sib := range mod.linkedSibs.Values() {
		list = append(list, sib.Id)
	}

	return list
}

func (mod *Module) addSibling(id *astral.Identity,
	action scheduler.ScheduledAction) {
	mod.linkedSibs.Set(id.String(), Sibling{
		Id:                 id,
		MaintainLinkAction: action,
	})
}
