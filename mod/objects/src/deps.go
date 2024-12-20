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
		var describers []any
		var finders []any

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

			if d, ok := m.(objects.Describer); ok {
				mod.AddDescriber(d)
				describers = append(describers, d)
			}

			if d, ok := m.(objects.Finder); ok {
				mod.AddFinder(d)
				finders = append(finders, d)
			}
		}

		if len(receivers) > 0 {
			mod.log.Logv(2, "receivers: %v"+strings.Repeat(", %s", len(receivers)-1), receivers...)
		}

		if len(holders) > 0 {
			mod.log.Logv(2, "holders: %v"+strings.Repeat(", %s", len(holders)-1), holders...)
		}

		if len(describers) > 0 {
			mod.log.Logv(2, "describers: %v"+strings.Repeat(", %s", len(describers)-1), describers...)
		}

		if len(finders) > 0 {
			mod.log.Logv(2, "finders: %v"+strings.Repeat(", %s", len(finders)-1), finders...)
		}
	}

	return
}
