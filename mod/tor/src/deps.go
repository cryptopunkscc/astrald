package tor

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

func (mod *Module) LoadDependencies(ctx *astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	modulePath := fmt.Sprintf(`/mod/%s`, tor.ModuleName)

	err = tree.BindPath(ctx, &mod.settings, mod.Tree.Root(), modulePath, true)
	if err != nil {
		return fmt.Errorf("tor module: bind settings: %w", err)
	}

	mod.Exonet.SetDialer("tor", mod)
	mod.Exonet.SetParser("tor", mod)
	mod.Exonet.SetUnpacker("tor", mod)
	mod.Nodes.AddResolver(mod)

	return nil
}
