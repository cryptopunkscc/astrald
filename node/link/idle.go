package link

import (
	"context"
	"errors"
	"time"
)

type Idle struct {
	link             *Link
	idleTimeout      time.Duration
	setIdleTimeoutCh chan time.Duration
}

func NewIdle(link *Link) *Idle {
	return &Idle{
		link:             link,
		setIdleTimeoutCh: make(chan time.Duration, 1),
		idleTimeout:      defaultIdleTimeout,
	}
}

func (l *Idle) IdleTimeout() time.Duration {
	return l.idleTimeout
}

// SetIdleTimeout sets the duriation of inactivity after which the link is closed. Set to 0 to never time out.
func (l *Idle) SetIdleTimeout(d time.Duration) error {
	select {
	case l.setIdleTimeoutCh <- d:
		return nil
	default:
		return errors.New("failed to set idle timeout")
	}
}

func (l *Idle) Run(ctx context.Context) error {
	for {
		if l.idleTimeout == 0 {
			select {
			case d := <-l.setIdleTimeoutCh:
				l.idleTimeout = d

			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			var left = l.idleTimeout - l.link.activity.Idle()
			if left <= 0 {
				return ErrIdleTimeout
			}

			select {
			case d := <-l.setIdleTimeoutCh:
				l.idleTimeout = d

			case <-time.After(left):

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
