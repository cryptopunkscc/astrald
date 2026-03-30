package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// RegisterHandler registers a handler and holds the registration until ctx is done or
// a connection failure occurs. It reconnects and re-registers automatically after a
// previously healthy registration session drops.
func (client *Client) RegisterHandler(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ch, err := client.queryCh(ctx, apphost.MethodRegisterHandler, query.Args{
			"endpoint": endpoint,
			"token":    authToken,
		})
		if err != nil {
			return err
		}

		// wait for registration ack
		if err = ch.Switch(channel.ExpectAck, channel.PassErrors); err != nil {
			ch.Close()
			return err
		}

		// hold the registration channel open as the lease; close it when context ends
		var done = make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				ch.Close()
			case <-done:
			}
		}()

		for {
			_, err = ch.Receive()
			if err != nil {
				break
			}
		}
		close(done)
		ch.Close()
	}
}

func RegisterHandler(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	return Default().RegisterHandler(ctx, endpoint, authToken)
}
