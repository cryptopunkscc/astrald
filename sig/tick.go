package sig

import "sync"

type Tick struct {
	mu sync.RWMutex
	ch chan struct{}
}

func (tick *Tick) Wait() <-chan struct{} {
	tick.mu.RLock()
	defer tick.mu.RUnlock()

	return tick.ch
}

func (tick *Tick) Tick() {
	tick.mu.Lock()
	defer tick.mu.Unlock()

	close(tick.ch)
	tick.ch = make(chan struct{})
}
