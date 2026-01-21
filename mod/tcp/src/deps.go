package tcp

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	ipmod "github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
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

	var modulePath = fmt.Sprintf(`/mod/%s`, tcp.ModuleName)

	var defaultListen = astral.Bool(mod.config.Listen)
	mod.listen, err = tree.Typed[*astral.Bool](mod.Tree.Bind(
		astral.NewContext(nil).WithIdentity(mod.node.Identity()),
		fmt.Sprintf(`%s/listen`, modulePath),
		&defaultListen,
		tree.TypedOnChange(mod.SwitchServer),
	))
	if err != nil {
		return err
	}

	var defaultDial = astral.Bool(mod.config.Dial)
	mod.dial, err = tree.Typed[*astral.Bool](mod.Tree.Bind(
		astral.NewContext(nil).WithIdentity(mod.node.Identity()),
		fmt.Sprintf(`%s/dial`, modulePath),
		&defaultDial,
		nil,
	))
	if err != nil {
		return err
	}

	mod.Exonet.SetDialer("tcp", mod)
	mod.Exonet.SetParser("tcp", mod)
	mod.Exonet.SetUnpacker("tcp", mod)
	mod.Nodes.AddResolver(mod)

	return
}
