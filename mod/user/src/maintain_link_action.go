package user

import (
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ scheduler.EventReceiver = &MaintainLinkAction{}
var _ scheduler.Task = &MaintainLinkAction{}

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
	retry, err := sig.NewRetry(time.Second, 15*time.Minute, 2)
	if err != nil {
		return err
	}

	a.mod.log.Log("starting to maintain link with %v", a.Target)
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

		switch {
		case count == 0:
			a.mod.log.Log("link to %v is broken, trying to reconnect", a.Target)
		case count > 0 && count%5 == 0:
			a.mod.log.Log("still trying to reconnect to %v (attempt %v)", a.Target, count)
		}

		resolve, err := a.mod.Nodes.ResolveEndpoints(ctx, a.Target)
		if err != nil {
			return nodes.ErrEndpointResolve
		}

		createStreamAction := a.mod.Nodes.NewCreateStreamAction(a.Target, sig.ChanToArray(resolve))
		scheduled, err := a.mod.Scheduler.Schedule(createStreamAction)
		if err != nil {
			return err
		}

		<-scheduled.Done()
		if scheduled.Err() != nil {
			count = <-retry.Retry()
			if count == 0 && a.actionRequired.Load() {
				count = 1
			}

			continue
		}

		retry.Reset()
		if count > 0 {
			a.mod.log.Log("link to %v restored after %v attempts", a.Target,
				count)
		} else if count < 0 {
			a.mod.log.Log("link to %v established", a.Target)
		}

		count = 0
		a.actionRequired.Store(false)
	}
}

func (a *MaintainLinkAction) ReceiveEvent(e *events.Event) {
	switch typed := e.Data.(type) {
	case *nodes.StreamClosedEvent:
		if !typed.RemoteIdentity.IsEqual(a.Target) || typed.StreamCount != 0 {
			return
		}

		if !a.actionRequired.Swap(true) {
			select {
			case a.wake <- struct{}{}:
			default:
			}
		}
	}
}
