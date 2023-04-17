package keepalive

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "net.keepalive"

type Loader struct{}

var log = _log.Tag(ModuleName)

func (Loader) Load(node node.Node) (node.Module, error) {
	mod := &Module{node: node}

	if err := node.ConfigStore().LoadYAML("keepalive", &mod.config); err != nil {
		log.Errorv(2, "error loading config: %s", err)
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
