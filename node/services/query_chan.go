package services

import "errors"

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

func (l QueryChan) Push(q *Query) error {
	select {
	case l <- q:
		return nil
	default:
		return ErrQueueOverflow
	}
}
