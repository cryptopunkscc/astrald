package sig

import (
	"context"
	"time"
)

// Idle returns a Signal that will be written to once when the Idler is idle for at least the provided duration.
func Idle(ctx context.Context, idler Idler, timeout time.Duration) Sig {
	sig := New()

	go func() {
		for {
			untilTimeout := timeout - idler.Idle()
			if untilTimeout <= 0 {
				sig <- struct{}{}
				return
			}

			select {
			case <-time.After(untilTimeout):
			case <-ctx.Done():
				return
			}
		}
	}()

	return sig
}
