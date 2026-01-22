package tcp

import (
	"fmt"
	"path"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	ipmod "github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type Deps struct {
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
	IP      ipmod.Module
	Tree    tree.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	modulePath := fmt.Sprintf(`/mod/%s`, tcp.ModuleName)

	var listen = astral.Bool(mod.config.Listen)
	mod.listen, err = tree.Bind[*astral.Bool](
		mod.Tree,
		path.Join(modulePath, "listen"),
		tree.WithOnChange(mod.SwitchServer),
		tree.WithDefaultValue(&listen),
	)
	if err != nil {
		return err
	}

	var dial = astral.Bool(mod.config.Dial)
	mod.dial, err = tree.Bind[*astral.Bool](
		mod.Tree,
		path.Join(modulePath, "dial"),
		tree.WithDefaultValue(&dial),
	)

	mod.Exonet.SetDialer("tcp", mod)
	mod.Exonet.SetParser("tcp", mod)
	mod.Exonet.SetUnpacker("tcp", mod)
	mod.Nodes.AddResolver(mod)

	return
}
