package relay

import (
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	mod.data, _ = modules.Load[data.Module](mod.node, data.ModuleName)
	mod.storage, _ = modules.Load[storage.Module](mod.node, storage.ModuleName)
	mod.keys, _ = modules.Load[keys.Module](mod.node, keys.ModuleName)

	return nil
}
