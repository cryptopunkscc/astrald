package sync

import (
	"context"
	"time"
)

// Timeout waits for an idler to be idle for at least timeout duration and returns nil. Returns an error if the
// context is done before timeout.
func Timeout(ctx context.Context, idler Idler, timeout time.Duration) *Signal {
	sig := &Signal{}

	go func() {
		if err := waitForTimeout(ctx, idler, timeout); err == nil {
			sig.Notify()
		}
	}()

	return sig
}

func waitForTimeout(ctx context.Context, idler Idler, timeout time.Duration) error {
	for {
		untilTimeout := timeout - idler.Idle()
		if untilTimeout <= 0 {
			return nil
		}

		select {
		case <-time.After(untilTimeout):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
