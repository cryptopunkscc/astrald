package gateway

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "gateway"

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *_log.Logger) (node.Module, error) {
	mod := &Module{
		node:        node,
		log:         log,
		config:      defaultConfig,
		dialer:      NewDialer(node),
		subscribers: make(map[string]*Subscriber),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	if i, ok := mod.node.Infra().(*core.CoreInfra); ok {
		i.SetDialer(NetworkName, mod.dialer)
		i.SetUnpacker(NetworkName, mod)
		i.SetParser(NetworkName, mod)
		i.AddEndpoints(mod)
	}

	log.Root().PushFormatFunc(func(v any) ([]_log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		var ops = make([]_log.Op, 0)

		if format, ok := log.Render(ep.gate); ok {
			ops = append(ops, format...)
		} else {
			ops = append(ops, _log.OpText{Text: ep.gate.String()})
		}

		ops = append(ops,
			_log.OpColor{Color: _log.White},
			_log.OpText{Text: ":"},
			_log.OpReset{},
		)

		if format, ok := log.Render(ep.target); ok {
			ops = append(ops, format...)
		} else {
			ops = append(ops, _log.OpText{Text: ep.gate.String()})
		}

		return ops, true
	})

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
