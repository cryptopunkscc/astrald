package sig

import (
	"context"
	"time"
)

var _ Waiter = &Timeout{}

type Timeout struct {
	sig chan struct{}
	t   time.Time
}

func SetTimeout(ctx context.Context, d time.Duration) *Timeout {
	a := &Timeout{
		sig: New(),
		t:   time.Now().Add(d),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(a.t)):
				if a.check() {
					close(a.sig)
					return
				}
			}
		}
	}()

	return a
}

// Defer makes sure the timeout will not happen for at least the provided duration.
func (t *Timeout) Defer(d time.Duration) {
	dt := time.Now().Add(d)

	if dt.After(t.t) {
		t.t = dt
	}
}

func (t *Timeout) Wait() <-chan struct{} {
	return t.sig
}

func (t *Timeout) check() bool {
	return !time.Now().Before(t.t)
}
