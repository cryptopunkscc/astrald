package auth

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
)

type Deps struct {
	Admin admin.Module
}

type Module struct {
	Deps
	config      Config
	node        astral.Node
	log         *log.Logger
	assets      resources.Resources
	authorizers sig.Set[auth.Authorizer]
}

func (mod *Module) Run(ctx context.Context) error {
	return nil
}

func (mod *Module) Authorize(identity *astral.Identity, action string, target astral.Object) bool {
	for _, a := range mod.authorizers.Clone() {
		if a.Authorize(identity, action, target) {
			name := reflect.TypeOf(a).String()
			if s, ok := a.(fmt.Stringer); ok {
				name = s.String()
			}

			var fmt = "%v allowed %v to %v"
			var vals = []any{name, identity, action}
			if target != nil {
				fmt += " on %v"
				vals = append(vals, target)
			}

			mod.log.Infov(2, fmt, vals...)
			return true
		}
	}

	if target == nil {
		mod.log.Logv(2, "denied %v to %v", identity, action)
	} else {
		mod.log.Logv(2, "denied %v to %v on %v", identity, action, target)
	}

	return false
}

func (mod *Module) AddAuthorizer(authorizer auth.Authorizer) error {
	return mod.authorizers.Add(authorizer)
}

func (mod *Module) Remove(authorizer auth.Authorizer) error {
	return mod.authorizers.Remove(authorizer)
}
