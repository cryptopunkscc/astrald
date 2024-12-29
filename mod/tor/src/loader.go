package tor

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"golang.org/x/net/proxy"
	"net"
)

const ModuleName = "tor"

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, logger *log.Logger) (core.Module, error) {
	mod := &Module{
		node:   node,
		log:    logger,
		assets: assets,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	mod.server = NewServer(mod)

	var baseDialer = &net.Dialer{Timeout: mod.config.DialTimeout}

	socksProxy, err := proxy.SOCKS5("tcp", mod.config.TorProxy, nil, baseDialer)
	if err != nil {
		return nil, err
	}

	if dialContext, ok := socksProxy.(proxy.ContextDialer); !ok {
		return nil, errors.New("type cast failed")
	} else {
		mod.proxy = dialContext
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
