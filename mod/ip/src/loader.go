package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node:   node,
		log:    l,
		config: Config{},
	}

	for _, addr := range mod.config.PublicEndpoints {
		ip, err := ip.ParseIP(addr)
		if err != nil {
			mod.log.Errorv(0,
				"ip module/Load invalid public endpoint IP: %v", addr)
			continue
		}
		mod.publicIPs = append(mod.publicIPs, ip)
	}

	_ = assets.LoadYAML(tcp.ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ip.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
