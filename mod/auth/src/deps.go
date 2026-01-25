package auth

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	if cnode, ok := mod.node.(*core.Node); ok {
		var added []any
		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}

			if a, ok := m.(auth.Authorizer); ok {
				mod.AddAuthorizer(a)
				added = append(added, m)
			}
		}
		if len(added) > 0 {
			mod.log.Logv(2, "authorizers: %v"+strings.Repeat(", %v", len(added)-1), added...)
		}
	}

	return
}
