package nodes

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"strings"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	node, ok := mod.node.(*core.Node)
	if !ok {
		return
	}

	var added []any
	for _, m := range node.Modules().Loaded() {
		if m == mod {
			continue
		}

		r, ok := m.(nodes.ServiceResolver)
		if !ok {
			continue
		}

		mod.AddServiceResolver(r)
		added = append(added, m)
	}
	
	if len(added) > 0 {
		mod.log.Logv(2, "service resolvers: %v"+strings.Repeat(", %v", len(added)-1), added...)
	}

	return
}
