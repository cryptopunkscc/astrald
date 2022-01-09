package linking

import (
	"context"
	"sync"
	"time"
)

type optimizer struct {
	*PeerOptimizer
	StartedAt time.Time
	EndAt     time.Time
	mu        sync.Mutex
}

func newOptimizer(peerOptimizer *PeerOptimizer, d time.Duration) *optimizer {
	return &optimizer{
		PeerOptimizer: peerOptimizer,
		StartedAt:     time.Now(),
		EndAt:         time.Now().Add(d),
	}
}

func (o *optimizer) optimize(d time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()

	target := time.Now().Add(d)

	if target.After(o.EndAt) {
		o.EndAt = target
	}
}

func (o optimizer) wait(ctx context.Context) error {
	for {
		o.mu.Lock()
		if time.Now().After(o.EndAt) {
			return nil
		}
		d := o.EndAt.Sub(time.Now())
		o.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
	}
}
