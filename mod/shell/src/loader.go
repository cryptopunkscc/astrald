package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		assets:     assets,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	_ = assets.LoadYAML(shell.ModuleName, &mod.config)

	type args struct {
		Name   astral.String    `query:"optional"`
		Target *astral.Identity `query:"optional"`
	}

	mod.root.AddOp("help", func(ctx astral.Context, stream shell.Stream, args args) error {
		stream.Write(term.Format("cool\n")[0])
		return nil
	})

	mod.AddRouteFunc("shell", mod.serve)

	return mod, err
}

func init() {
	if err := core.RegisterModule(shell.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
