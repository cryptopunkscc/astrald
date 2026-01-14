package sig

import "context"

type PhaseSplitter[T any] struct {
	before   <-chan T
	after    <-chan T
	boundary Signal
}

func (p *PhaseSplitter[T]) Before() <-chan T          { return p.before }
func (p *PhaseSplitter[T]) After() <-chan T           { return p.after }
func (p *PhaseSplitter[T]) Boundary() <-chan struct{} { return p.boundary.Done() }

// Done makes PhaseSplitter compatible with Signal / On / OnCtx utilities.
func (p *PhaseSplitter[T]) Done() <-chan struct{} { return p.boundary.Done() }

// NewPhaseSplitter splits a source stream into two phases separated by a boundary marker.
func NewPhaseSplitter[T any](
	ctx context.Context,
	src <-chan T,
	isBoundary func(T) bool,
) *PhaseSplitter[T] {
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

	return &PhaseSplitter[T]{
		before:   before,
		after:    after,
		boundary: Sig(boundaryCh),
	}
}
