package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	modmedia "github.com/cryptopunkscc/astrald/mod/media"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:   node,
		log:    log,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(modmedia.ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(modmedia.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
