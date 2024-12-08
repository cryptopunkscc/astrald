package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"strconv"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
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

	l.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(*Endpoint)
		if !ok {
			return nil, false
		}

		var ops = make([]log.Op, 0)

		ip := ep.ip.String()
		if ep.ip.IsIPv6() {
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
	if err := core.RegisterModule(tcp.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
