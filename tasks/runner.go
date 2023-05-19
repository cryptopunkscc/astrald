package tasks

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Runner is an interface that wraps the basic Run method.
// Run runs a task within the provided context and returns an error.
type Runner interface {
	Run(context.Context) error
}

type RunFunc func(context.Context) error

type GroupRunner struct {
	runners     []Runner
	DoneHandler func(Runner, error)
}

type RunError struct {
	err    error
	runner Runner
	id     int
}

type RunFuncAdapter struct {
	RunFunc func(context.Context) error
}

func (r RunFuncAdapter) Run(ctx context.Context) error {
	if r.RunFunc == nil {
		return errors.New("run function is nil")
	}
	return r.RunFunc(ctx)
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

// Run runs all runners concurrently and returns after all runners are done. If DoneHandler is set, it
// will be called after a runner returns. Returns nil or context error.
func (g *GroupRunner) Run(ctx context.Context) (err error) {
	var wg sync.WaitGroup

	for _, runner := range g.runners {
		runner := runner

		if runner == nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err = runner.Run(ctx)
			if g.DoneHandler != nil {
				g.DoneHandler(runner, err)
			}
		}()
	}

	wg.Wait()
	return ctx.Err()
}
