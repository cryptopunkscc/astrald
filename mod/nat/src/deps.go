package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

// LoadDependencies injects required dependencies via the core injector.
func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return err
	}

	return nil
}
