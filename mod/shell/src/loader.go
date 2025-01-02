package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(shell.ModuleName, &mod.config)

	mod.root.AddOp("ops", func(ctx astral.Context, q shell.Query) error {
		t, err := shell.AcceptTerminal(q)
		if err != nil {
			return err
		}
		defer t.Close()

		ops := mod.root.Tree()
		slices.Sort(ops)
		for _, o := range ops {
			t.Print((*astral.String8)(&o), term.Newline{})
		}

		return nil
	})

	mod.root.AddOp("shell", mod.opShell)

	return mod, err
}

func init() {
	if err := core.RegisterModule(shell.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
