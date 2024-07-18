package fwd

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tor"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.tcp, err = core.Load[tcp.Module](mod.node, tcp.ModuleName)
	if err != nil {
		return err
	}

	mod.tor, err = core.Load[tor.Module](mod.node, tor.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
