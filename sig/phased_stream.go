package sig

import "context"

type PhasedStream[T any] struct {
	before   <-chan T
	after    <-chan T
	boundary Signal
}

func (p *PhasedStream[T]) Before() <-chan T          { return p.before }
func (p *PhasedStream[T]) After() <-chan T           { return p.after }
func (p *PhasedStream[T]) Boundary() <-chan struct{} { return p.boundary.Done() }

// Done makes PhasedStream compatible with Signal / On / OnCtx utilities.
func (p *PhasedStream[T]) Done() <-chan struct{} { return p.boundary.Done() }

// NewPhasedStream splits a source stream into two phases separated by a boundary marker.
func NewPhasedStream[T any](
	ctx context.Context,
	src <-chan T,
	isBoundary func(T) bool,
) *PhasedStream[T] {
	const buf = 64

	before := make(chan T, buf)
	after := make(chan T, buf)
	boundaryCh := make(chan struct{})

	go func() {
		defer close(after)

		boundarySignaled := false
		beforeClosed := false

		signalBoundary := func() {
			if boundarySignaled {
				return
			}
			boundarySignaled = true
			close(boundaryCh)
		}

		closeBefore := func() {
			if beforeClosed {
				return
			}
			beforeClosed = true
			close(before)
		}

		inBefore := true

		for {
			select {
			case <-ctx.Done():
				// Ensure snapshot consumers are unblocked.
				if inBefore {
					signalBoundary()
					closeBefore()
				}
				return

			case v, ok := <-src:
				if !ok {
					if inBefore {
						signalBoundary()
						closeBefore()
					}
					return
				}

				if inBefore && isBoundary(v) {
					// Phase transition
					inBefore = false
					signalBoundary()
					closeBefore()
					continue
				}

				if inBefore {
					select {
					case before <- v:
					case <-ctx.Done():
						signalBoundary()
						closeBefore()
						return
					}
					continue
				}

				select {
				case after <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return &PhasedStream[T]{
		before:   before,
		after:    after,
		boundary: Sig(boundaryCh),
	}
}
