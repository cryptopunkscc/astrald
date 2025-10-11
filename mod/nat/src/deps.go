package nat

import (
	"github.com/cryptopunkscc/astrald/core"
)

// LoadDependencies injects required dependencies via the core injector.
func (mod *Module) LoadDependencies() (err error) {
	if err = core.Inject(mod.node, &mod.Deps); err != nil {
		return err
	}

	return nil
}
