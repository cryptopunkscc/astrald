package tcpfwd

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "net.tcpfwd"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	mod := &Module{
		node:   node,
		assets: assets,
		config: defaultConfig,
		log:    log,
	}

	_ = assets.LoadYAML("tcpfwd", &mod.config)

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
