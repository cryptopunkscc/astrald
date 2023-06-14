package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/link/ctl"
	"io"
)

type Control struct {
	link *Link
}

func NewControl(link *Link) *Control {
	return &Control{link: link}
}

func (c *Control) Run(ctx context.Context) error {
	for {
		msg, err := c.link.ctl.ReadMessage()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case ctl.QueryMessage:
			if err := c.link.handleQueryMessage(ctx, msg); err != nil {
				return err
			}

		case ctl.DropMessage:
			if err := c.link.onDrop(msg.Port()); err != nil {
				return err
			}

		case ctl.CloseMessage:
			if err := c.link.onClose(); err != nil {
				return err
			}
			return io.EOF

		case ctl.PingMessage:
			if err := c.link.onPing(msg.Port()); err != nil {
				return err
			}

		default:
			return errors.New("unknown control message")
		}

		// check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}
