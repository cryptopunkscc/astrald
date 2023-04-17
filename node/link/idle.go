package link

import (
	"context"
	"errors"
	"time"
)

type Idle struct {
	link         *Link
	timeout      time.Duration
	setTimeoutCh chan time.Duration
}

func NewIdle(link *Link) *Idle {
	return &Idle{
		link:         link,
		setTimeoutCh: make(chan time.Duration, 1),
		timeout:      defaultIdleTimeout,
	}
}

func (idle *Idle) Timeout() time.Duration {
	return idle.timeout
}

// SetTimeout sets the duriation of inactivity after which the link is closed. Set to 0 to never time out.
func (idle *Idle) SetTimeout(d time.Duration) error {
	select {
	case idle.setTimeoutCh <- d:
		return nil
	default:
		return errors.New("failed to set idle timeout")
	}
}

func (idle *Idle) Run(ctx context.Context) error {
	for {
		if idle.timeout == 0 {
			select {
			case d := <-idle.setTimeoutCh:
				idle.timeout = d

			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			var left = idle.timeout - idle.link.activity.Idle()
			if left <= 0 {
				idle.link.CloseWithError(ErrIdleTimeout)
				return ErrIdleTimeout
			}

			select {
			case d := <-idle.setTimeoutCh:
				idle.timeout = d

			case <-time.After(left):

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
