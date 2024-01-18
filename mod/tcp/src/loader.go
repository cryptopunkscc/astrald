package tcp

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"strconv"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, l *log.Logger) (modules.Module, error) {
	mod := &Module{
		node:   node,
		log:    l,
		config: defaultConfig,
	}

	_ = assets.LoadYAML(tcp.ModuleName, &mod.config)

	// Parse public endpoints
	for _, pe := range mod.config.PublicEndpoints {
		endpoint, err := Parse(pe)
		if err != nil {
			l.Error("error parsing public endpoint \"%s\": %s", pe, err)
			continue
		}

		mod.publicEndpoints = append(mod.publicEndpoints, endpoint)
	}

	node.Infra().SetDialer("tcp", mod)
	node.Infra().SetParser("tcp", mod)
	node.Infra().SetUnpacker("tcp", mod)
	node.Infra().SetDialer("inet", mod)
	node.Infra().SetParser("inet", mod)
	node.Infra().SetUnpacker("inet", mod)
	node.Infra().AddEndpoints(mod)

	l.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		var ops = make([]log.Op, 0)

		ip := ep.ip.String()
		if ep.ver == ipv6 {
			ip = "[" + ip + "]"
		}

		ops = append(ops,
			log.OpColor{Color: log.Cyan},
			log.OpText{Text: ip},
			log.OpReset{},
		)

		if ep.port != 0 {
			ops = append(ops,
				log.OpColor{Color: log.White},
				log.OpText{Text: ":"},
				log.OpReset{},
				log.OpColor{Color: log.Cyan},
				log.OpText{Text: strconv.Itoa(int(ep.port))},
				log.OpReset{},
			)
		}

		return ops, true
	})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(tcp.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
