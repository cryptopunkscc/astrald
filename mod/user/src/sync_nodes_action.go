package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type SyncNodesAction struct {
	remoteIdentity *astral.Identity
	mod            *Module
}

func (mod *Module) NewSyncNodesAction(remoteIdentity *astral.
	Identity) user.SyncNodesAction {
	return &SyncNodesAction{
		remoteIdentity: remoteIdentity,
		mod:            mod,
	}
}

func (a *SyncNodesAction) String() string {
	return "user.sync_nodes_action"
}

func (a *SyncNodesAction) Run(ctx *astral.Context) error {
	err := a.mod.SyncApps(ctx, a.remoteIdentity)
	if err != nil {
		a.mod.log.Error("error syncing apps with %v: %v", a.remoteIdentity, err)
	}

	err = a.mod.SyncAlias(ctx, a.remoteIdentity)
	if err != nil {
		a.mod.log.Error("error syncing alias of %v: %v", a.remoteIdentity, err)
	}

	err = a.mod.SyncAssets(ctx, a.remoteIdentity)
	if err != nil {
		a.mod.log.Error("error syncing assets of %v: %v", a.remoteIdentity, err)
	}

	return nil
}
