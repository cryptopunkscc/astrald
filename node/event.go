package node

import "sync"

type Eventer interface {
	Event() string
}

type FutureEvent struct {
	wait  chan struct{}
	event Eventer
	next  *FutureEvent
	mu    sync.Mutex
}

func NewFutureEvent() *FutureEvent {
	return &FutureEvent{wait: make(chan struct{})}
}

func (e *FutureEvent) Wait() <-chan struct{} {
	return e.wait
}

func (e *FutureEvent) Event() Eventer {
	return e.event
}

func (e *FutureEvent) Next() *FutureEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.next == nil {
		e.next = NewFutureEvent()
	}

	return e.next
}

func (e *FutureEvent) done(ev Eventer) *FutureEvent {
	defer close(e.wait)
	e.event = ev
	return e.Next()
}
