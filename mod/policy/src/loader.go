package policy

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/policy"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *_log.Logger) (node.Module, error) {
	mod := &Module{
		node:     node,
		log:      log,
		config:   defaultConfig,
		policies: make(map[*RunningPolicy]struct{}),
	}

	_ = assets.LoadYAML(policy.ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(policy.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
