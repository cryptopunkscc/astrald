package user

import (
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ scheduler.EventReceiver = &MaintainLinkAction{}
var _ scheduler.Action = &MaintainLinkAction{}

// MaintainLinkAction attempts to maintain a link to a target node indefinitely.
// triggers:
// - ensure_connectivity_action
type MaintainLinkAction struct {
	mod            *Module
	Target         *astral.Identity
	wake           chan struct{}
	actionRequired atomic.Bool
}

func (mod *Module) NewMaintainLinkAction(target *astral.
	Identity) user.MaintainLinkAction {
	return &MaintainLinkAction{
		mod:    mod,
		Target: target,
		wake:   make(chan struct{}, 1),
	}
}

func (a *MaintainLinkAction) String() string { return "nodes.maintain_link" }

func (a *MaintainLinkAction) Run(ctx *astral.Context) error {
	// check if we are already linked

	count := -1
	a.actionRequired.Store(!a.mod.Nodes.IsLinked(a.Target))
	for {
		for !a.actionRequired.Load() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-a.wake:
			}
		}

		if count == 0 {
			a.mod.log.Log("link to %v is broken, trying to reconnect", a.Target)
		}

		if count > 0 && count%5 == 0 {
			a.mod.log.Log("still trying to reconnect to %v (attempt %v)",
				a.Target, count)
		}

		createStreamAction := a.mod.Nodes.NewCreateStreamAction(a.Target.String(), "", "")
		scheduled := a.mod.Scheduler.Schedule(ctx, createStreamAction)
		<-scheduled.Done()
		if scheduled.Err() != nil {
			if count < 0 {
				count = 0
			}

			count++
			time.Sleep(5 * time.Second)
			continue
			// FIXME: backoff struct (sig.Backoff)
		}

		// Success path
		if count > 0 {
			a.mod.log.Log("link to %v restored after %v attempts", a.Target,
				count)
		} else if count < 0 {
			a.mod.log.Log("link to %v established", a.Target)
		}

		count = 0 // reset for future real outages
		a.actionRequired.Store(false)
	}
}

func (a *MaintainLinkAction) ReceiveEvent(e *events.Event) {
	switch typed := e.Data.(type) {
	case *nodes.StreamClosedEvent:
		if !typed.RemoteIdentity.IsEqual(a.Target) {
			return
		}

		if typed.StreamCount == 0 {
			a.actionRequired.Store(true)
			a.wake <- struct{}{}
		}

	}
}
