package sig

import (
	"context"
	"sync"
	"time"
)

func At(at *time.Time, lock sync.Locker, fn func()) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		for {
			if lock != nil {
				lock.Lock()
			}
			var t = *at
			if lock != nil {
				lock.Unlock()
			}

			if t.IsZero() {
				return
			}

			if time.Now().After(t) {
				fn()
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(t)):
			}
		}
	}()

	return cancel
}
