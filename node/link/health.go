package link

import (
	"context"
	"time"
)

type health struct {
	link *CoreLink
	sig  chan struct{}
}

func newHealth(link *CoreLink) *health {
	return &health{
		link: link,
		sig:  make(chan struct{}, 1),
	}
}

func (h *health) Run(ctx context.Context) error {
	for {
		select {
		case <-h.sig:
			_, err := h.link.control.Ping()
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()

			case <-h.link.Done():
				return h.link.err

			case <-time.After(time.Second):
			}

		case <-ctx.Done():
			return ctx.Err()

		case <-h.link.Done():
			return h.link.err
		}
	}
}

func (h *health) Check() {
	select {
	case h.sig <- struct{}{}:
	default:
	}
}