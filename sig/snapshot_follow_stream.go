package sig

import "context"

// SnapshotFollowStream produces a stream with strict snapshot-follow semantics stream can be splitted using PhaseSplitter.
func SnapshotFollowStream[T any](
	ctx context.Context,
	snapshot []T,
	updates <-chan T,
	flush T,
) <-chan T {
	out := make(chan T, 16)

	go func() {
		defer close(out)

		for _, v := range snapshot {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}

		select {
		case out <- flush:
		case <-ctx.Done():
			return
		}

		// If updates is nil, return early instead of blocking forever
		if updates == nil {
			return
		}

		for {
			select {
			case v, ok := <-updates:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
