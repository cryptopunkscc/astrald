package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
}

type Module struct {
	Deps
	config   Config
	node     astral.Node
	log      *log.Logger
	assets   resources.Resources
	handlers sig.Map[auth.Action, []auth.Handler]
}

func (mod *Module) Run(ctx *astral.Context) error {
	return nil
}

func (mod *Module) Authorize(ctx *astral.Context, identity *astral.Identity, action auth.Action, target astral.Object) bool {
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	for _, h := range mod.get(action) {
		if h(ctx, identity, target) {
			mod.log.Logv(3, "allowed %v to %v", identity, action)
			return true
		}
	}

	if target == nil {
		mod.log.Logv(3, "denied %v to %v", identity, action)
	} else {
		mod.log.Logv(3, "denied %v to %v on %v [%v]", identity, action, target, target.ObjectType())
	}

	return false
}

func (mod *Module) AddAuthorizer(action auth.Action, handlers ...auth.Handler) {
	mod.handlers.Set(action, append(mod.get(action), handlers...))
}

func (mod *Module) get(action auth.Action) []auth.Handler {
	h, _ := mod.handlers.Get(action)
	return h
}
