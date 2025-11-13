package scheduler

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

type ScheduledTask struct {
	scheduledAt astral.Time
	task        scheduler.Task
	done        chan struct{}
	cancel      context.CancelCauseFunc
	cancelOnce  *sync.Once
	err         error
	state       atomic.Int64
}

func NewScheduledTask(task scheduler.Task, cancelCauseFunc context.CancelCauseFunc) *ScheduledTask {
	return &ScheduledTask{
		task:        task,
		scheduledAt: astral.Now(),
		done:        make(chan struct{}),
		cancel:      cancelCauseFunc,
		cancelOnce:  &sync.Once{},
	}
}

// Task returns the scheduled task
func (h *ScheduledTask) Task() scheduler.Task {
	return h.task
}

// Err returns the error returned by the task. Should only be called after the task is done.
func (h *ScheduledTask) Err() error {
	return h.err
}

func (h *ScheduledTask) CancelWithError(err error) {
	h.err = err
	h.cancel(err)
	return
}

func (h *ScheduledTask) Done() <-chan struct{} {
	return h.done
}

// called externally
func (h *ScheduledTask) Cancel() {
	h.CancelWithError(context.Canceled)
	h.close()
	return
}

// called by scheduler
func (h *ScheduledTask) close() {
	h.cancelOnce.Do(func() {
		close(h.done)
	})

	return
}

func (h ScheduledTask) ScheduledAt() astral.Time {
	return h.scheduledAt
}

func (h *ScheduledTask) State() scheduler.State {
	return scheduler.State(h.state.Load())
}

func (h *ScheduledTask) setState(state scheduler.State) {
	h.state.Store(int64(state))
}
