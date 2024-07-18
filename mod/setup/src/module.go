package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/setup"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ setup.Module = &Module{}

type Module struct {
	*routers.PathRouter
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets

	user     user.Module
	keys     keys.Module
	relay    relay.Module
	apphost  apphost.Module
	presence presence.Module
	dir      dir.Module

	inviteService *InviteService
}

func (mod *Module) Run(ctx context.Context) error {
	mod.inviteService = NewInviteService(mod, func(identity id.Identity) bool {
		return true
	})

	return tasks.Group(
		&SetupService{Module: mod},
		mod.inviteService,
	).Run(ctx)
}

func (mod *Module) Invite(ctx context.Context, userID id.Identity, nodeID id.Identity) error {
	return mod.inviteService.Invite(ctx, userID, nodeID)
}

func (mod *Module) OnPendingAd(ad presence.PendingAd) {
	if mod.needsSetup() {
		ad.AddFlag(presence.SetupFlag)
	}
}

func (mod *Module) needsSetup() bool {
	return mod.user.UserID().IsZero()
}
