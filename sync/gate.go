package sync

import (
	"sync"
)

// Gate can be closed or open. Wait doesn't block when the gate is open and blocks otherwise.
type Gate struct {
	ch   chan struct{}
	mu   sync.Mutex
	open bool
}

// Wait returns a channel that's closed when the gate is open.
func (gate *Gate) Wait() <-chan struct{} {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	if gate.ch == nil {
		gate.close()
	}

	return gate.ch
}

// Open sets gate state to open and closes the channel returned by Wait
func (gate *Gate) Open() {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	if gate.ch == nil {
		gate.ch = make(chan struct{})
	}

	if !gate.open {
		gate.open = true
		close(gate.ch)
	}
}

// Close sets gate state to closed and makes a new channel for Wait
func (gate *Gate) Close() {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	gate.close()
}

func (gate *Gate) close() {
	if gate.ch == nil {
		gate.ch = make(chan struct{})
	}

	if gate.open {
		gate.open = false
		gate.ch = make(chan struct{})
	}
}
