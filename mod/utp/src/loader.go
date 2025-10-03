package utp

import (
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

	// Parse public endpoints
	for _, pe := range mod.config.PublicEndpoints {
		endpoint, err := utp.ParseEndpoint(pe)
		if err != nil {
			l.Error("error parsing public endpoint \"%v\": %v", pe, err)
			continue
		}

		mod.publicEndpoints = append(mod.publicEndpoints, endpoint)
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(utp.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
