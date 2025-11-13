package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

// Ensure Module struct implements the public scheduler.Module interface
var _ scheduler.Module = (*Module)(nil)

// Module is the concrete implementation of the scheduler module.
type Module struct {
	Deps

	ctx    *astral.Context
	node   astral.Node
	log    *log.Logger
	assets resources.Resources

	queue sig.Set[scheduler.ScheduledTask]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	<-ctx.Done()
	return nil
}

// Schedule schedules a task to be executed once all dependencies are done.
func (mod *Module) Schedule(task scheduler.Task, deps ...scheduler.Done) (_ scheduler.ScheduledTask, err error) {
	// check if the scheduler is running
	if !mod.isRunning() {
		mod.log.Errorv(2, "cannot schedule (not running): %v", task)
		return nil, scheduler.ErrNotRunning
	}

	// check if the task is nil
	if task == nil {
		return nil, scheduler.ErrTaskIsNil
	}

	// create a scheduled task
	sTask := NewScheduledTask(task)

	// add the task to the queue
	// err is always nil, because duplicates are impossible here
	_ = mod.queue.Add(sTask)

	// spawn the task runner goroutine
	go func() {
		defer mod.queue.Remove(sTask)

		// wait for all dependencies
		for _, d := range deps {
			select {
			case <-sTask.Done():
				return
			case <-d.Done():
			}
		}

		// run the task within the context of the scheduler module
		err = sTask.Run(mod.ctx)

		// log on error
		if err != nil {
			mod.log.Errorv(2, "task %v: %v", task, err)
		}

		// release all releasable dependencies
		for _, d := range deps {
			if rh, ok := d.(scheduler.Releaser); ok && rh != nil {
				rh.Release()
			}
		}
	}()

	return sTask, nil
}

// isRunning returns true if the scheduler module is running.
func (mod *Module) isRunning() bool {
	if mod.ctx == nil {
		return false
	}
	select {
	case <-mod.ctx.Done():
		return false
	default:
	}
	return true
}

func (mod *Module) String() string {
	return scheduler.ModuleName
}
