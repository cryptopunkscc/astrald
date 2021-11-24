package sync

import "sync"

type Signal struct {
	ch chan struct{}
	mu sync.Mutex
}

func (sig *Signal) Wait() <-chan struct{} {
	sig.mu.Lock()
	defer sig.mu.Unlock()

	// reset the channel if necessary
	if sig.ch == nil {
		sig.reset()
	}

	return sig.ch
}

func (sig *Signal) Notify() {
	sig.mu.Lock()
	defer sig.mu.Unlock()

	if sig.ch != nil {
		close(sig.ch)
	}
	sig.reset()
}

func (sig *Signal) reset() {
	sig.ch = make(chan struct{})
}
