package objects

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(objects.ModuleName, NewAdmin(mod))

	if cnode, ok := mod.node.(*core.Node); ok {
		var receivers []any
		var holders []any

		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}

			if r, ok := m.(objects.Receiver); ok {
				mod.AddReceiver(r)
				receivers = append(receivers, r)
			}

			if h, ok := m.(objects.Holder); ok {
				mod.AddHolder(h)
				holders = append(holders, h)
			}
		}

		if len(receivers) > 0 {
			mod.log.Logv(2, "object receivers: %v"+strings.Repeat(", %s", len(receivers)-1), receivers...)
		}

		if len(holders) > 0 {
			mod.log.Logv(2, "object holders: %v"+strings.Repeat(", %s", len(holders)-1), holders...)
		}
	}

	return
}
