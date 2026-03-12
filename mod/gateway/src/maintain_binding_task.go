package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ scheduler.Task = &MaintainBindingTask{}

type MaintainBindingTask struct {
	mod        *Module
	GatewayID  *astral.Identity
	Visibility gateway.Visibility
}

func (mod *Module) NewMaintainBindingTask(gatewayID *astral.Identity, visibility gateway.Visibility) *MaintainBindingTask {
	return &MaintainBindingTask{
		mod:        mod,
		GatewayID:  gatewayID,
		Visibility: visibility,
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

	client := gatewayClient.New(task.GatewayID, astrald.Default())

	count := -1
	for {
		switch {
		case count == 0:
			task.mod.log.Log("binding to %v lost, rebinding", task.GatewayID)
		case count > 0 && count%5 == 0:
			task.mod.log.Log("still trying to bind to %v (attempt %v)", task.GatewayID, count)
		}

		socket, err := client.Bind(ctx.IncludeZone(astral.ZoneNetwork), task.Visibility)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case count = <-retry.Retry():
			}
			continue
		}

		retry.Reset()
		if count > 0 {
			task.mod.log.Log("rebound to %v after %v attempts", task.GatewayID, count)
		} else if count < 0 {
			task.mod.log.Log("bound to %v", task.GatewayID)
		}
		count = 0

		err = task.mod.newSocketPool(ctx, socket).Run()
		if err != nil {
			task.mod.log.Error("rebinding to %v due to: %v", task.GatewayID, err)
		}
	}
}
