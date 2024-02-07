package presence

import (
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.tcp, err = modules.Load[tcp.Module](mod.node, tcp.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = modules.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
