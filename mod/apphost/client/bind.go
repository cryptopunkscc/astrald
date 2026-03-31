package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// Bind establishes a status channel with the host. It blocks until ctx is done or the channel
// drops. A non-nil error returned while ctx is still live signals a connection loss.
func (client *Client) Bind(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	ch, err := client.queryCh(ctx, apphost.MethodBind, query.Args{
		"endpoint": endpoint,
		"token":    authToken,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	if err = ch.Switch(channel.ExpectAck, channel.PassErrors); err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)
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

	if ctx.Err() != nil {
		return nil
	}

	return err
}

func Bind(ctx *astral.Context, endpoint string, authToken astral.Nonce) error {
	return Default().Bind(ctx, endpoint, authToken)
}
