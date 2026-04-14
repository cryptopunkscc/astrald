package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func (mod *Module) ActiveLocalAppContracts() ([]*auth.SignedContract, error) {
	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity()).ExcludeZone(astral.ZoneNetwork)
	return mod.Auth.SignedContracts().
		WithSubject(mod.node.Identity()).
		WithAction(&apphost.HostForAction{}).
		Find(ctx)
}
