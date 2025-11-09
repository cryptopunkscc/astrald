package user

import (
	"fmt"
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
	}
}

func (a *MaintainLinkAction) String() string { return "nodes.maintain_link" }

func (a *MaintainLinkAction) Run(ctx *astral.Context) error {
	// check if we are already linked
	a.actionRequired.Store(!a.mod.Nodes.IsLinked(a.Target))
	for {
		for !a.actionRequired.Load() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-a.wake:
			}
		}

		createStreamAction := a.mod.Nodes.NewCreateStreamAction(a.Target.String(), "", "")
		scheduled := a.mod.Scheduler.Schedule(ctx, createStreamAction)
		<-scheduled.Done()
		if scheduled.Err() != nil {
			time.Sleep(3 * time.Second)
			fmt.Println("retrying")
			continue
			// FIXME: backoff struct (sig.Backoff)
		}

		a.mod.log.Log("added sibling %v", a.Target)
		a.actionRequired.Store(false)
	}

}

func (a *MaintainLinkAction) ReceiveEvent(e *events.Event) {
	switch typed := e.Data.(type) {
	case *nodes.StreamClosedEvent:
		// FIXME: never happens
		fmt.Println("stream closed")
		a.actionRequired.Store(typed.StreamCount == 0)
		a.wake <- struct{}{}
	}
}
