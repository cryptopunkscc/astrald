package tasks

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Runner is an interface that wraps the basic Run method.
// Run runs a task within the provided context and returns an error.
type Runner interface {
	Run(context.Context) error
}

type GroupRunner struct {
	runners []Runner
}

type RunError struct {
	err    error
	runner Runner
	id     int
}

func (e RunError) Error() string {
	return fmt.Sprintf("runner %s (#%d) failed: %s", reflect.TypeOf(e.runner), e.id, e.err.Error())
}

func (e RunError) Unwrap() error {
	return e.err
}

func (e RunError) Runner() Runner {
	return e.runner
}

func (e RunError) ID() int {
	return e.id
}

func Group(runners ...Runner) *GroupRunner {
	return &GroupRunner{runners: runners}
}

// Run all runners concurrently. If any of the runners returns an error, context for all other runners will
// be canceled and the first error is returned wrapped in a RunError. Returns nil if all runners return
// without an error.
func (g *GroupRunner) Run(ctx context.Context) (err error) {
	var runCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var ch = make(chan error, len(g.runners))

	for i, r := range g.runners {
		r := r
		i := i

		if r == nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Run(runCtx); err != nil {
				ch <- RunError{err: err, runner: r, id: i}
				cancel()
			}
		}()
	}

	wg.Wait()

	select {
	case err = <-ch:
	default:
	}

	return
}
