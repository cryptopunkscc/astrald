package services

import (
	"context"
	"sync"
)

// ServiceFeed manages the lifecycle of a service with state tracking and subscriber management.
// It provides thread-safe state updates and automatic broadcasting to subscribers.
type ServiceFeed struct {
	mu          sync.RWMutex
	service     *Service
	subscribers map[chan ServiceChange]context.CancelFunc
}

// NewServiceFeed creates a new ServiceFeed with optional initial service state.
func NewServiceFeed(initialService *Service) *ServiceFeed {
	return &ServiceFeed{
		service:     initialService,
		subscribers: make(map[chan ServiceChange]context.CancelFunc),
	}
}

// SetService updates the service state and notifies all subscribers.
// Pass nil to disable the service.
func (t *ServiceFeed) SetService(svc *Service) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.service = svc

	// Broadcast to all subscribers
	change := ServiceChange{
		Enabled: svc != nil,
	}
	if svc != nil {
		change.Service = *svc
	}

	for ch := range t.subscribers {
		select {
		case ch <- change:
		default:
			// Skip slow subscribers (non-blocking)
		}
	}
}

// DisableService is a convenience method to disable the service.
// It's equivalent to SetService(nil).
func (t *ServiceFeed) DisableService() {
	t.SetService(nil)
}

// GetSnapshot returns the current service state.
// Returns nil if the service is disabled.
func (t *ServiceFeed) GetSnapshot() *Service {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.service
}

// Subscribe creates a new subscription channel for service changes.
// The channel will be automatically cleaned up when the context is cancelled.
// Returns a buffered channel that receives ServiceChange events.
func (t *ServiceFeed) Subscribe(ctx context.Context) <-chan ServiceChange {
	ch := make(chan ServiceChange, 16)

	t.mu.Lock()
	// Create a cancel function for cleanup
	ctx, cancel := context.WithCancel(ctx)
	t.subscribers[ch] = cancel
	t.mu.Unlock()

	// Start cleanup goroutine
	go func() {
		<-ctx.Done()
		t.unsubscribe(ch)
	}()

	return ch
}

// unsubscribe removes a subscriber channel and closes it.
func (t *ServiceFeed) unsubscribe(ch chan ServiceChange) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if cancel, exists := t.subscribers[ch]; exists {
		cancel()
		delete(t.subscribers, ch)
		close(ch)
	}
}

// Close cleans up all subscribers and resources.
func (t *ServiceFeed) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for ch, cancel := range t.subscribers {
		cancel()
		close(ch)
	}
	t.subscribers = make(map[chan ServiceChange]context.CancelFunc)
}
