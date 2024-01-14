package policy

import (
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	mod.relay, _ = modules.Load[relay.Module](mod.node, relay.ModuleName)

	return nil
}
