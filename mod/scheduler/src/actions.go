package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// WaitableAction wraps an Action to expose a Done() signal when finished.
type WaitableAction struct {
	scheduler.Action
	done chan struct{}
	err  error
}

// NewWaitable wraps an action, returning a waitable wrapper.
func NewWaitable(a scheduler.Action) *WaitableAction {
	return &WaitableAction{Action: a, done: make(chan struct{})}
}

func (a *WaitableAction) Done() <-chan struct{} { return a.done }

func (a *WaitableAction) Error() error { return a.err }

func (a *WaitableAction) Run(ctx *astral.Context) error {
	defer close(a.done)
	// call the embedded action's Run, not ourselves
	err := a.Action.Run(ctx)
	a.err = err
	return err
}
