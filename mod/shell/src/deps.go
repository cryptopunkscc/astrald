package shell

import (
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"strings"
)

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies() (err error) {
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

			if s, ok := m.(shell.HasScope); ok {
				mod.root.AddScope(term.Stringify(s), s.Scope())
				added = append(added, m)
			}
		}
		if len(added) > 0 {
			mod.log.Logv(2, "shell scopes: %v"+strings.Repeat(", %v", len(added)-1), added...)
		}
	}

	return
}
