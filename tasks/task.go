package tasks

import (
	"context"
	"sync/atomic"
	"time"
)

// Task is wraps a ResultRunner interface and provides additional functionality.
type Task[T any] struct {
	runner    ResultRunner[T]
	done      chan struct{}
	cancel    chan struct{}
	runtime   time.Duration
	startTime time.Time
	result    T
	err       error
	ran       atomic.Bool
	canceled  atomic.Bool
}

// New instantiates a new Task using the provided ResultRunner
func New[T any](runner ResultRunner[T]) *Task[T] {
	return &Task[T]{
		runner: runner,
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
		err:    nil,
	}
}

// NewFunc instantiates a new Task using the provided function
func NewFunc[T any](runnerFunc TaskFunc[T]) *Task[T] {
	return New[T](&funcAdapter[T]{TaskFunc: runnerFunc})
}

// Run runs the underlying ResultRunner. After it's done, the channel returned by Done() will be closed and
// Result() and Err() methods can be used to read the values returned by the ResultRunner.
func (t *Task[T]) Run(ctx context.Context) error {
	if !t.ran.CompareAndSwap(false, true) {
		return ErrTaskAlreadyRun
	}

	defer close(t.done)

	select {
	case <-t.cancel:
		t.err = context.Canceled
		return t.err

	default:
	}

	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-t.cancel:
			cancel()
		case <-taskCtx.Done():
		}
	}()

	t.startTime = time.Now().Round(0)
	t.result, t.err = t.runner.Run(taskCtx)
	t.runtime = time.Since(t.startTime)

	return t.err
}

// Cancel cancels the task. If it has not been run yet, its error is set to context.Canceled and the underlying
// ResultRunner will not be run. If the task is already running, its context will be canceled.
func (t *Task[T]) Cancel() {
	if t.canceled.CompareAndSwap(false, true) {
		close(t.cancel)
	}
}

// Done returns a channel which will close when task finishes running.
func (t *Task[T]) Done() <-chan struct{} {
	return t.done
}

// Result waits for the task to finish and returns the result of the task. If the task was canceled it returns
// zero value.
func (t *Task[T]) Result() T {
	select {
	case <-t.done:
	case <-t.cancel:
	}
	return t.result
}

// Err waits for the task to finish and returns the error returned by the task. If the task was canceled it returns
// context.Canceled.
func (t *Task[T]) Err() error {
	select {
	case <-t.done:
	case <-t.cancel:
	}
	return t.err
}

// Runtime waits for the task to finish and returns its running duration. If the task was canceled it returns zero.
func (t *Task[T]) Runtime() time.Duration {
	select {
	case <-t.done:
	case <-t.cancel:
	}
	return t.runtime
}
