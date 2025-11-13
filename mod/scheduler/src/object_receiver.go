package scheduler

import (
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

func (mod *Module) ReceiveObject(drop objects.Drop) (err error) {
	switch o := drop.Object().(type) {
	case *events.Event:
		for _, task := range mod.queue.Clone() {
			// skip non-running tasks
			if task.State() != scheduler.StateRunning {
				continue
			}

			// propagate the event to the task if supported
			if receiver, ok := task.Task().(scheduler.EventReceiver); ok {
				receiver.ReceiveEvent(o)
			}
		}
	}

	return nil
}
