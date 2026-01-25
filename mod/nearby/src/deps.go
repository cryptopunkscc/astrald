package nearby

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/ether"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Ether   ether.Module
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
	Shell   shell.Module
	TCP     tcp.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Dir.AddResolver(mod)
	mod.Nodes.AddResolver(mod)

	if cnode, ok := mod.node.(*core.Node); ok {
		var composers []any
		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}
			if a, ok := m.(nearby.Composer); ok {
				mod.AddStatusComposer(a)
				composers = append(composers, a)
			}
		}

		if mod.composers.Count() > 0 {
			mod.log.Logv(2, "composers: %v"+strings.Repeat(", %v", len(composers)-1), composers...)
		}
	}

	return nil
}
