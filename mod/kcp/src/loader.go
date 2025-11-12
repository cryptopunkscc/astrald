package kcp

import (
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node:   node,
		log:    l,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(kcpmod.ModuleName, &mod.config)

	for _, addr := range mod.config.Endpoints {
		addr, _ = strings.CutPrefix(addr, fmt.Sprintf("%s:", kcpmod.ModuleName))

		endpoint, err := kcpmod.ParseEndpoint(addr)
		if err != nil {
			mod.log.Errorv(0, "kcp module/Load invalid endpoint: %v", addr)
		}

		mod.configEndpoints = append(mod.configEndpoints, endpoint)
	}

	mod.ops.AddStruct(mod, "Op")

	return mod, nil
}

func init() {
	if err := core.RegisterModule(kcpmod.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
