package ip

import (
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies() (err error) {
	return core.Inject(mod.node, &mod.Deps)
}
