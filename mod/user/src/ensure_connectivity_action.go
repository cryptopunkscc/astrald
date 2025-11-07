package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
)

// EnsureConnectivityAction ensures that we have at least one stream with all siblings.
// triggers:
// - on node startup
// - nodes.StreamClosedEvent
type EnsureConnectivityAction struct {
	mod *Module
}

func (mod *Module) NewEnsureConnectivityAction() user.EnsureConnectivityAction {
	return &EnsureConnectivityAction{mod: mod}
}

func (a *EnsureConnectivityAction) Run(ctx *astral.Context) error {
	ac := a.mod.ActiveContract()
	if ac == nil {
		return nil
	}

	// We are missing connection with certain nodes.
	for _, node := range a.mod.LocalSwarm() {
		if a.mod.Nodes.IsLinked(node) {
			continue
		}

		createStreamAction := a.mod.Nodes.NewCreateStreamAction(node.String(), "", "")
		scheduledAction := a.mod.Scheduler.Schedule(ctx, createStreamAction)
		go func() {
			<-scheduledAction.Wait()
			a.mod.linkedSibs.Set(node.String(), node)

			err := a.mod.SyncApps(ctx, node)
			if err != nil {
				a.mod.log.Error("error syncing apps with %v: %v", node, err)
			}

			err = a.mod.SyncAlias(ctx, node)
			if err != nil {
				a.mod.log.Error("error syncing alias of %v: %v", node, err)
			}

			err = a.mod.SyncAssets(ctx, node)
			if err != nil {
				a.mod.log.Error("error syncing assets of %v: %v", node, err)
			}
		}()
	}

	return nil
}

func (a *EnsureConnectivityAction) String() string {
	return "user.ensure_connectivity_action"
}
