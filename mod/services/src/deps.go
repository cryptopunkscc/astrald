package services

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/services"
)

type Deps struct {
}

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

			if d, ok := m.(services.ServiceDiscoverer); ok {
				mod.AddServiceDiscoverer(d)
			}
		}
	}

	return nil
}
