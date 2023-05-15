package keepalive

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "net.keepalive"
const configName = "keepalive"

type Loader struct{}

var log = _log.Tag(ModuleName)

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{node: node}

	configStore.LoadYAML(configName, &mod.config)

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
