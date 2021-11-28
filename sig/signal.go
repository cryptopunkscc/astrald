package sig

import (
	"context"
)

var _ Waiter = Signal(nil)

type Signal <-chan struct{}

func New() chan struct{} {
	return make(chan struct{}, 1)
}

func (s Signal) Wait() <-chan struct{} {
	return s
}

func (s Signal) WaitContext(ctx context.Context) error {
	select {
	case <-s:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
