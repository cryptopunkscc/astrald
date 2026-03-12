package gateway

import (
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ scheduler.Task = &MaintainBindingTask{}
var _ scheduler.EventReceiver = &MaintainBindingTask{}

type MaintainBindingTask struct {
	mod            *Module
	GatewayID      *astral.Identity
	Visibility     gateway.Visibility
	wake           chan struct{}
	actionRequired atomic.Bool
}

func (mod *Module) NewMaintainBindingTask(gatewayID *astral.Identity, visibility gateway.Visibility) *MaintainBindingTask {
	return &MaintainBindingTask{
		mod:        mod,
		GatewayID:  gatewayID,
		Visibility: visibility,
		wake:       make(chan struct{}, 1),
	}
}

func (task *MaintainBindingTask) String() string {
	return "maintain_binding_task"
}

func (task *MaintainBindingTask) Run(ctx *astral.Context) error {
	task.mod.log.Log("starting to maintain binding to %v", task.GatewayID)

	retry, err := sig.NewRetry(time.Second, 15*time.Minute, 2)
	if err != nil {
		return err
	}

	count := -1
	task.actionRequired.Store(true)

	for {
		for !task.actionRequired.Load() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-task.wake:
			}
		}

		switch {
		case count == 0:
			task.mod.log.Log("binding to %v lost, rebinding", task.GatewayID)
		case count > 0 && count%5 == 0:
			task.mod.log.Log("still trying to bind to %v (attempt %v)", task.GatewayID, count)
		}

		client := gatewayClient.New(task.GatewayID, astrald.Default())
		socket, bindErr := client.Bind(ctx.IncludeZone(astral.ZoneNetwork), task.Visibility)
		if bindErr != nil {
			count = <-retry.Retry()
			continue
		}

		retry.Reset()
		if count > 0 {
			task.mod.log.Log("rebound to %v after %v attempts", task.GatewayID, count)
		} else if count < 0 {
			task.mod.log.Log("bound to %v", task.GatewayID)
		}
		count = 0
		task.actionRequired.Store(false)

		task.maintainSocketConnections(ctx, socket)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		// socket dead — trigger rebind
		task.actionRequired.Store(true)
		select {
		case task.wake <- struct{}{}:
		default:
		}
	}
}

func (task *MaintainBindingTask) maintainSocketConnections(ctx *astral.Context, socket *gateway.Socket) {
	if err := newSocketPool(task.mod, socket).Run(ctx); err == errSocketUnreachable {
		task.mod.log.Log("gateway socket %v unreachable, will rebind", socket.Endpoint)
	}
}

func (task *MaintainBindingTask) ReceiveEvent(e *events.Event) {
	switch typed := e.Data.(type) {
	case *nodes.StreamClosedEvent:
		if !typed.RemoteIdentity.IsEqual(task.GatewayID) || typed.StreamCount != 0 {
			return
		}
		if !task.actionRequired.Swap(true) {
			select {
			case task.wake <- struct{}{}:
			default:
			}
		}
	}
}
