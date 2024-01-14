package presence

import (
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.tcp, err = modules.Load[tcp.Module](mod.node, tcp.ModuleName)

	return err
}
