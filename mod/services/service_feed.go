package services

import (
	"context"
	"sync"
)

// ServiceFeed is a pure broadcast primitive for ServiceChange events.
//
// It does NOT store or expose any current service state and provides no snapshot APIs.
// Producers are responsible for maintaining current service state themselves.
type ServiceFeed struct {
	mu          sync.Mutex
	subscribers map[chan ServiceChange]struct{}
	closed      bool
}

func NewServiceFeed() *ServiceFeed {
	return &ServiceFeed{
		subscribers: make(map[chan ServiceChange]struct{}),
	}
}

// Subscribe registers a new subscriber and returns a channel of ServiceChange.
//
// When ctx.Done() fires, the subscription is removed and the channel is closed.
func (f *ServiceFeed) Subscribe(ctx context.Context) <-chan ServiceChange {
	ch := make(chan ServiceChange, 16)

	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		close(ch)
		return ch
	}
	f.subscribers[ch] = struct{}{}
	f.mu.Unlock()

	go func() {
		<-ctx.Done()
		f.unsubscribe(ch)
	}()

	return ch
}

func (f *ServiceFeed) unsubscribe(ch chan ServiceChange) {
	f.mu.Lock()
	_, ok := f.subscribers[ch]
	if ok {
		delete(f.subscribers, ch)
	}
	f.mu.Unlock()

	if ok {
		close(ch)
	}
}

// Publish broadcasts a ServiceChange to all current subscribers.
//
// Non-blocking per subscriber: if a subscriber channel is full, the event may be dropped
// for that subscriber.
func (f *ServiceFeed) Publish(change ServiceChange) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}

	// Snapshot subscriber list under lock.
	subs := make([]chan ServiceChange, 0, len(f.subscribers))
	for ch := range f.subscribers {
		subs = append(subs, ch)
	}
	f.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- change:
		default:
			// Drop for slow subscribers.
		}
	}
}

// Close cancels and removes all subscribers and closes their channels.
func (f *ServiceFeed) Close() {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	f.closed = true

	subs := make([]chan ServiceChange, 0, len(f.subscribers))
	for ch := range f.subscribers {
		subs = append(subs, ch)
	}
	f.subscribers = make(map[chan ServiceChange]struct{})
	f.mu.Unlock()

	for _, ch := range subs {
		close(ch)
	}
}
