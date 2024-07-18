package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	_log "github.com/cryptopunkscc/astrald/log"
)

const ModuleName = "gateway"

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *_log.Logger) (core.Module, error) {
	mod := &Module{
		node:        node,
		log:         log,
		PathRouter:  routers.NewPathRouter(node.Identity(), false),
		config:      defaultConfig,
		dialer:      NewDialer(node),
		subscribers: make(map[string]*Subscriber),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

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
