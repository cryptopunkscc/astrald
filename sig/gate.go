package sig

import (
	"sync"
)

// Gate can be closed or open. Wait doesn't block when the gate is open and blocks otherwise.
type Gate struct {
	ch      chan struct{}
	mu      sync.Mutex
	open    bool
	flipped bool
}

// Wait returns a channel that's closed when the gate is open.
func (gate *Gate) Wait() <-chan struct{} {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	if gate.ch == nil {
		gate.setClosed()
	}

	return gate.ch
}

// Open sets gate state to open and closes the channel returned by Wait
func (gate *Gate) Open() {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	if gate.flipped {
		gate.setClosed()
	} else {
		gate.setOpen()
	}
}

// Close sets gate state to closed and makes a new channel for Wait
func (gate *Gate) Close() {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	if gate.flipped {
		gate.setOpen()
	} else {
		gate.setClosed()
	}
}

func (gate *Gate) Flip() *Gate {
	gate.mu.Lock()
	defer gate.mu.Unlock()

	gate.flipped = !gate.flipped

	if gate.open {
		gate.setClosed()
	} else {
		gate.setOpen()
	}

	return gate
}

func (gate *Gate) setOpen() {
	if gate.ch == nil {
		gate.ch = make(chan struct{})
	}

	if !gate.open {
		gate.open = true
		close(gate.ch)
	}
}

func (gate *Gate) setClosed() {
	if gate.ch == nil {
		gate.ch = make(chan struct{})
	}

	if gate.open {
		gate.open = false
		gate.ch = make(chan struct{})
	}
}
