package services

import "context"

type PhasedStream[T any] struct {
	before   <-chan T
	after    <-chan T
	boundary <-chan struct{}
}

func (p *PhasedStream[T]) Before() <-chan T          { return p.before }
func (p *PhasedStream[T]) After() <-chan T           { return p.after }
func (p *PhasedStream[T]) Boundary() <-chan struct{} { return p.boundary }

func NewPhasedStream[T any](
	ctx context.Context,
	src <-chan T,
	isBoundary func(T) bool,
) *PhasedStream[T] {
	before := make(chan T)
	after := make(chan T)
	boundary := make(chan struct{})

	go func() {
		defer close(before)
		defer close(after)
		defer close(boundary)

		inBefore := true

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-src:
				if !ok {
					return
				}

				if inBefore && isBoundary(v) {
					inBefore = false
					close(boundary)
					continue
				}

				if inBefore {
					before <- v
				} else {
					after <- v
				}
			}
		}
	}()

	return &PhasedStream[T]{
		before:   before,
		after:    after,
		boundary: boundary,
	}
}
