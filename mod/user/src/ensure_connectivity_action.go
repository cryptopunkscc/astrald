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

		maintainLinkAction := a.mod.NewMaintainLinkAction(node)

		scheduledAction := a.mod.Scheduler.Schedule(ctx, maintainLinkAction)
		a.mod.addSibling(node, scheduledAction.Cancel)
	}

	return nil
}

func (a *EnsureConnectivityAction) String() string {
	return "user.ensure_connectivity_action"
}
