package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

const ModuleName = "gateway"

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:        node,
		log:         log,
		PathRouter:  routers.NewPathRouter(node.Identity(), false),
		config:      defaultConfig,
		dialer:      NewDialer(node),
		subscribers: make(map[string]*Subscriber),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	term.SetTranslateFunc(func(e *gateway.Endpoint) astral.Object {
		return &term.ColorString{
			Color: term.HighlightColor,
			Text:  astral.String32(e.String()),
		}
	})

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
