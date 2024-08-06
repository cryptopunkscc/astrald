package objects

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(objects.ModuleName, NewAdmin(mod))
	mod.Auth.AddAuthorizer(mod)

	if cnode, ok := mod.node.(*core.Node); ok {
		for _, m := range cnode.Modules().Loaded() {
			if r, ok := m.(objects.Receiver); ok {
				var name = fmt.Sprintf("%s", m)
				mod.log.Logv(2, "auto-added %v as object receiver", name)
				mod.AddReceiver(r)
			}

			if h, ok := m.(objects.Holder); ok {
				var name = fmt.Sprintf("%s", m)
				mod.log.Logv(2, "auto-added %v as object holder", name)
				mod.AddHolder(h)
			}
		}
	}

	return
}
