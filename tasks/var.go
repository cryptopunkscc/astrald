package tasks

import (
	"context"
	"errors"
)

// ErrQueueOverflow - task could not be added to the queue, because the queue is full
var ErrQueueOverflow = errors.New("task queue overflow")

// ErrAlreadyRunning - cannot run the scheduler because it's already running
var ErrAlreadyRunning = errors.New("already running")

// ErrTaskAlreadyRun - task has been run already
var ErrTaskAlreadyRun = errors.New("already run")

// Runner is an interface that wraps the basic Run method.
// Run runs a task within the provided context and returns an error.
type Runner interface {
	Run(context.Context) error
}

// ResultRunner is an iterface that wraps a Run method that returns a value in addition to an error.
type ResultRunner[T any] interface {
	Run(context.Context) (T, error)
}

// TaskFunc is a function that can be used as a tasks (see NewFunc).
type TaskFunc[T any] func(context.Context) (T, error)

type funcAdapter[T any] struct {
	TaskFunc[T]
}

func (a *funcAdapter[T]) Run(ctx context.Context) (T, error) {
	return a.TaskFunc(ctx)
}
