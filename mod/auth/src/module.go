package auth

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
)

type Deps struct {
}

type Module struct {
	Deps
	config      Config
	node        astral.Node
	log         *log.Logger
	assets      resources.Resources
	authorizers sig.Set[auth.Authorizer]
}

func (mod *Module) Run(ctx *astral.Context) error {
	return nil
}

func (mod *Module) Authorize(identity *astral.Identity, action auth.Action, target astral.Object) bool {
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	for _, a := range mod.authorizers.Clone() {
		if a.Authorize(identity, action, target) {
			name := reflect.TypeOf(a).String()
			if s, ok := a.(fmt.Stringer); ok {
				name = s.String()
			}

			var fmt = "%v allowed %v to %v"
			var vals = []any{name, identity, action}
			if target != nil {
				fmt += " on %v [%v]"
				vals = append(vals, target, target.ObjectType())
			}

			mod.log.Infov(3, fmt, vals...)
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

func (mod *Module) AddAuthorizer(authorizer auth.Authorizer) error {
	return mod.authorizers.Add(authorizer)
}

func (mod *Module) Remove(authorizer auth.Authorizer) error {
	return mod.authorizers.Remove(authorizer)
}
