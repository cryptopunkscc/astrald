package keepalive

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
)

const ModuleName = "net.keepalive"
const configName = "keepalive"

type Loader struct{}

var log = _log.Tag(ModuleName)

func (Loader) Load(node node.Node, configStore config.Store) (node.Module, error) {
	mod := &Module{node: node}

	configStore.LoadYAML(configName, &mod.config)

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
