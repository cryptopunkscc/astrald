package ether

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/ether"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Admin   admin.Module
	Objects objects.Module
	Keys    keys.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(ether.ModuleName, NewAdmin(mod))
	mod.Objects.Blueprints().Add(
		&ether.Broadcast{},
		&ether.SignedBroadcast{},
		&ether.EventBroadcastReceived{},
	)

	return
}
