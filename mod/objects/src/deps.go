package objects

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	if cnode, ok := mod.node.(*core.Node); ok {
		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}

			if d, ok := m.(objects.Describer); ok {
				mod.AddDescriber(d)
			}

			if d, ok := m.(objects.Purger); ok {
				mod.AddPurger(d)
			}

			if d, ok := m.(objects.Searcher); ok {
				mod.AddSearcher(d)
			}

			if d, ok := m.(objects.Finder); ok {
				mod.AddFinder(d)
			}

			if h, ok := m.(objects.Holder); ok {
				mod.AddHolder(h)
			}

			if r, ok := m.(objects.Receiver); ok {
				mod.AddReceiver(r)
			}
		}

	}

	return
}
