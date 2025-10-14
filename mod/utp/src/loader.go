package utp

import (
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node:   node,
		log:    l,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(utp.ModuleName, &mod.config)

	for _, addr := range mod.config.Endpoints {
		addr, _ = strings.CutPrefix(addr, fmt.Sprintf("%s:", utp.ModuleName))

		endpoint, err := utp.ParseEndpoint(addr)
		if err != nil {
			mod.log.Errorv(0, "tcp module/Load invalid endpoint: %v", addr)
		}

		mod.configEndpoints = append(mod.configEndpoints, endpoint)
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(utp.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
