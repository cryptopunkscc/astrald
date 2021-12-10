package event

import (
	"context"
	"github.com/cryptopunkscc/astrald/sig"
)

type Queue struct {
	*sig.Queue
}

func NewQueue() *Queue {
	return &Queue{
		Queue: &sig.Queue{},
	}
}

func (q *Queue) Push(eventer Eventer) *Queue {
	return &Queue{
		Queue: q.Queue.Push(eventer),
	}
}

func (q *Queue) Next() *Queue {
	return &Queue{Queue: q.Queue.Next()}
}

func (q *Queue) Data() Eventer {
	return q.Queue.Data().(Eventer)
}

func (q *Queue) Follow(ctx context.Context) <-chan Eventer {
	ch := make(chan Eventer)

	go func() {
		defer close(ch)
		for i := range q.Queue.Follow(ctx) {
			ch <- i.(Eventer)
		}
	}()

	return ch
}
