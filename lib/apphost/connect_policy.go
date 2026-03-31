package apphost

import (
	"time"

	"github.com/cryptopunkscc/astrald/sig"
)

// ConnectPolicy is called after each failed connect attempt.
type ConnectPolicy func(attempt int, err error) (time.Duration, error)

// defaultPolicy retries with exponential backoff forever.
func defaultPolicy() ConnectPolicy {
	r, _ := sig.NewRetry(250*time.Millisecond, 10*time.Second, 2)
	return func(attempt int, err error) (time.Duration, error) {
		if attempt == 1 {
			r.Reset()
			return 0, nil
		}
		d := r.NextDelay()
		go func() { <-r.Retry() }() // consume to advance delay; fires after d, same as the loop's sleep
		return d, nil
	}
}
