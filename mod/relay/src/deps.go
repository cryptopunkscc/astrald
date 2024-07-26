package relay

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(relay.ModuleName, NewAdmin(mod))

	mod.Objects.AddPrototypes(relay.CertDesc{})

	return
}
