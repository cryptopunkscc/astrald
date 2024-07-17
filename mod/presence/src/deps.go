package presence

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.tcp, err = core.Load[tcp.Module](mod.node, tcp.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	mod.keys, err = core.Load[keys.Module](mod.node, keys.ModuleName)
	if err != nil {
		return err
	}

	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return err
	}

	mod.auth, err = core.Load[auth.Module](mod.node, auth.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
