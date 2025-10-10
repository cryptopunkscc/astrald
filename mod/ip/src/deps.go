package ip

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return err
	}

	if cnode, ok := mod.node.(*core.Node); ok {
		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}

			if d, ok := m.(ip.CandidateFinder); ok {
				mod.AddFinder(d)
			}

		}
	}

	return nil
}
