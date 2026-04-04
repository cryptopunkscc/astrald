// lib/astrald/retry_policy.go
package astrald

import (
	"math"
	"time"
)

// RetryPolicy governs wait time and whether to retry on consecutive failures.
// attempt is zero-indexed and resets to 0 after each success.
type RetryPolicy interface {
	Next(attempt int, err error) (time.Duration, bool)
}

type backoff struct {
	min, max    time.Duration
	factor      float64
	maxAttempts int // 0 = infinite
}

func Backoff(min, max time.Duration, factor float64) RetryPolicy {
	return &backoff{min: min, max: max, factor: factor}
}

func LimitedBackoff(min, max time.Duration, factor float64, maxAttempts int) RetryPolicy {
	return &backoff{min: min, max: max, factor: factor, maxAttempts: maxAttempts}
}

func (b *backoff) Next(attempt int, _ error) (time.Duration, bool) {
	if b.maxAttempts > 0 && attempt >= b.maxAttempts {
		return 0, false
	}
	d := time.Duration(float64(b.min) * math.Pow(b.factor, float64(attempt)))
	if d > b.max {
		d = b.max
	}
	return d, true
}

type noReconnect struct{}

func NoReconnect() RetryPolicy                            { return noReconnect{} }
func (noReconnect) Next(int, error) (time.Duration, bool) { return 0, false }
