package events

import (
	"context"
	"github.com/cryptopunkscc/astrald/sig"
)

// HandlerRunner wraps the Handle function in a tasks.Runner interface
type HandlerRunner[EventType Event] struct {
	*Queue
	Func func(EventType) error
}

// Handle will subscribe to the Queue for the duration of the context and will invoke fn for every
// element that matches fn's argument type. If fn returns an error, Handle stops and retruns the error.
func Handle[EventType Event](ctx context.Context, q *Queue, fn func(EventType) error) error {
	return sig.Handle(ctx, q.getQueue(), fn)
}

// Runner creates a new instance of HandlerRunner
func Runner[EventType Event](q *Queue, fn func(EventType) error) *HandlerRunner[EventType] {
	return &HandlerRunner[EventType]{
		Queue: q,
		Func:  fn,
	}
}

func (h *HandlerRunner[EventType]) Run(ctx context.Context) error {
	return Handle(ctx, h.Queue, h.Func)
}
