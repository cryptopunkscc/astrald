package sig

import (
	"context"
	"sync"
)

// Switch manages a togglable background task with at-most-one instance running.
type Switch struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	done   <-chan struct{}
}

// Run starts the task if not already running. The task receives a cancellable
// child context derived from ctx. Returns true if started, false if already running.
func (s *Switch) Run(ctx context.Context, task func(context.Context)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running() {
		return false
	}

	ctx, s.cancel = context.WithCancel(ctx)
	done := make(chan struct{})
	s.done = done

	go func() {
		defer close(done)
		task(ctx)
	}()

	return true
}

// Stop cancels the running task. Returns true if stopped, false if not running.
func (s *Switch) Stop() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running() {
		return false
	}

	s.cancel()
	s.cancel = nil

	return true
}

// Set starts or stops the task based on the boolean.
func (s *Switch) Set(ctx context.Context, on bool, task func(context.Context)) {
	if on {
		s.Run(ctx, task)
	} else {
		s.Stop()
	}
}

// Running returns true if a task is currently active.
func (s *Switch) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.running()
}

func (s *Switch) running() bool {
	if s.done == nil {
		return false
	}

	select {
	case <-s.done:
		return false
	default:
		return true
	}
}
