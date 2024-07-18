package auth

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
	"strings"
)

type Module struct {
	config      Config
	node        astral.Node
	log         *log.Logger
	assets      resources.Resources
	authorizers sig.Set[auth.Authorizer]
}

func (mod *Module) Run(ctx context.Context) error {
	return nil
}

func (mod *Module) Authorize(identity id.Identity, action string, args ...any) bool {
	for _, a := range mod.authorizers.Clone() {
		if a.Authorize(identity, action, args...) {
			name := reflect.TypeOf(a).String()
			if s, ok := a.(fmt.Stringer); ok {
				name = s.String()
			}

			var fmt = "allowed %v to %v" + strings.Repeat(" %v", len(args)) + " by %v"
			var vals = []any{identity, action}
			vals = append(vals, args...)
			vals = append(vals, name)

			mod.log.Infov(2, fmt, vals...)
			return true
		}
	}

	mod.log.Logv(2, "denied %v to %v", identity, action)

	return false
}

func (mod *Module) AddAuthorizer(authorizer auth.Authorizer) error {
	return mod.authorizers.Add(authorizer)
}

func (mod *Module) Remove(authorizer auth.Authorizer) error {
	return mod.authorizers.Remove(authorizer)
}
