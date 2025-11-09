package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ scheduler.Action = &MaintainLinkAction{}

// MaintainLinkAction attempts to maintain a link to a target node indefinitely.
// triggers:
// - ensure_connectivity_action
type MaintainLinkAction struct {
	mod    *Module
	Target *astral.Identity
}

func (mod *Module) NewMaintainLinkAction(target *astral.
	Identity) user.MaintainLinkAction {
	return &MaintainLinkAction{
		mod:    mod,
		Target: target,
	}
}

func (a *MaintainLinkAction) String() string { return "nodes.maintain_link" }

func (a *MaintainLinkAction) Run(ctx *astral.Context) error {
	backoff := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// spawn one CreateStreamAction attempt
		createStreamAction := a.mod.Nodes.NewCreateStreamAction(a.Target.String(), "", "")
		scheduled := a.mod.Scheduler.Schedule(ctx, createStreamAction)
		<-scheduled.Done()

		info, err := createStreamAction.Result()
		if err != nil {
			delay := min((1<<backoff)*time.Second, 15*time.Minute)
			backoff = min(backoff+1, 32)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		a.mod.log.Log("maintain_link_action for %v", a.Target)
		// NOTE: could be part of stream created event handler
		err = a.mod.SyncApps(ctx, info.RemoteIdentity)
		if err != nil {
			a.mod.log.Error("error syncing apps with %v: %v", info.RemoteIdentity, err)
		}

		err = a.mod.SyncAlias(ctx, info.RemoteIdentity)
		if err != nil {
			a.mod.log.Error("error syncing alias of %v: %v", info.RemoteIdentity, err)
		}

		err = a.mod.SyncAssets(ctx, info.RemoteIdentity)
		if err != nil {
			a.mod.log.Error("error syncing assets of %v: %v", info.RemoteIdentity, err)
		}

		return nil
	}
}
