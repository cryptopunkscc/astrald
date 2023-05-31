package services

import (
	"context"
	"errors"
)

type QueryChan chan *Query

func NewQueryChan(queueSize int) QueryChan {
	return make(chan *Query, queueSize)
}

func (l QueryChan) Accept() (*Conn, error) {
	q, ok := <-l
	if !ok {
		return nil, errors.New("channel closed")
	}
	return q.Accept()
}

func (l QueryChan) Push(ctx context.Context, q *Query) error {
	select {
	case l <- q:
		return nil
	case <-ctx.Done():
		return ctx.Err()

	default:
		return ErrQueueOverflow
	}
}
