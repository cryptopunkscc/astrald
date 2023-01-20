package link

import (
	"context"
	"errors"
	"time"
)

func (l *Link) IdleTimeout() time.Duration {
	return l.idleTimeout
}

// SetIdleTimeout sets the duriation of inactivity after which the link is closed. Set to 0 to never time out.
func (l *Link) SetIdleTimeout(d time.Duration) error {
	select {
	case l.setIdleTimeoutCh <- d:
		return nil
	default:
		return errors.New("failed to set idle timeout")
	}
}

func (l *Link) monitorIdle(ctx context.Context) error {
	for {
		if l.idleTimeout == 0 {
			select {
			case d := <-l.setIdleTimeoutCh:
				l.idleTimeout = d

			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			var left = l.idleTimeout - l.Idle()
			if left <= 0 {
				l.err = ErrIdleTimeout
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
