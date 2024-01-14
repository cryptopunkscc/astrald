package tor

import (
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"golang.org/x/net/proxy"
	"net"
)

const ModuleName = "tor"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, logger *log.Logger) (modules.Module, error) {
	mod := &Module{
		node:   node,
		log:    logger,
		assets: assets,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	node.Infra().SetDialer("tor", mod)
	node.Infra().SetParser("tor", mod)
	node.Infra().SetUnpacker("tor", mod)
	node.Infra().AddEndpoints(mod)

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

	logger.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		return []log.Op{
			log.OpColor{Color: log.Cyan},
			log.OpText{Text: ep.String()},
			log.OpReset{},
		}, true
	})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
