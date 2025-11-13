package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// FIXME: write tests

// Schedule enqueues an task for processing by launching a goroutine that
// waits for dependencies, runs the task, and then releases resources.
// It is safe for concurrent use.
func (mod *Module) Schedule(
	ctx *astral.Context,
	task scheduler.Task,
	deps ...scheduler.Done,
) (scheduledAction scheduler.ScheduledTask, err error) {
	// check if the scheduler is running
	if !mod.isRunning() {
		err = scheduler.ErrNotRunning
		mod.log.Errorv(2, "dropped %v (%v)", task, err)
		return
	}

	// check if the task is valid
	if task == nil {
		return nil, scheduler.ErrTaskIsNil
	}

	// create a subcontext for the task
	taskCtx, cancel := ctx.WithCancelCause()
	sTask := NewScheduledTask(task, cancel)

	// add the task to the queue
	// err is always nil, because duplicates are impossible here
	_ = mod.queue.Add(sTask)

	go func() {
		defer sTask.close()
		defer mod.queue.Remove(sTask)

		// wait for all dependencies
		for _, d := range deps {
			select {
			case <-taskCtx.Done():
				sTask.err = taskCtx.Err()
				return
			case <-d.Done():
			}
		}

		// run the task
		sTask.setState(scheduler.StateRunning)
		sTask.err = sTask.task.Run(taskCtx)
		sTask.setState(scheduler.StateDone)

		// log on error
		if sTask.err != nil {
			mod.log.Errorv(2, "task %v: %v", task, sTask.err)
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
