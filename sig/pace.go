package sig

import (
	"context"
	"time"
)

func Pace(ctx context.Context, interval time.Duration, burst int) Signal {
	sig := New()

	for i := 0; i < burst; i++ {
		sig <- struct{}{}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				select {
				case sig <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return sig
}
