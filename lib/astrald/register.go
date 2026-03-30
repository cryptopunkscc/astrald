package astrald

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

type RetryHandler func(ctx *astral.Context, attempt int, err error) bool

type RegisterOption func(*registerOptions)

type registerOptions struct {
	retryHandler      RetryHandler
	disconnectHandler func(error)
	reconnectHandler  func()
}

func WithRetryHandler(h RetryHandler) RegisterOption {
	return func(o *registerOptions) { o.retryHandler = h }
}

func WithDisconnectHandler(h func(error)) RegisterOption {
	return func(o *registerOptions) { o.disconnectHandler = h }
}

func WithReconnectHandler(h func()) RegisterOption {
	return func(o *registerOptions) { o.reconnectHandler = h }
}

// Register registers the listener with apphost and holds the registration until ctx is done,
// the listener is closed, or retrying is stopped by a custom RetryHandler. Retries automatically
// on failure using exponential backoff.
func (l *Listener) Register(ctx *astral.Context, opts ...RegisterOption) error {
	var options registerOptions
	for _, o := range opts {
		o(&options)
	}

	child, cancel := ctx.WithCancel()
	defer cancel()

	go func() {
		select {
		case <-l.Done():
			cancel()
		case <-child.Done():
		}
	}()

	err := registerHandler(child, l.client, l.Endpoint(), l.AuthToken(), options)

	select {
	case <-ctx.Done():
		return nil
	case <-l.Done():
		return nil
	default:
		return err
	}
}

func registerHandler(ctx *astral.Context, client *Client, endpoint string, authToken astral.Nonce, opts registerOptions) error {
	retryFn := opts.retryHandler
	if retryFn == nil {
		retryFn = newDefaultRetryHandler()
	}

	disconnected := false
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ch, err := client.QueryChannel(ctx, apphost.MethodRegisterHandler, query.Args{
			"endpoint": endpoint,
			"token":    authToken,
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !disconnected {
				if opts.disconnectHandler != nil {
					opts.disconnectHandler(err)
				}
				disconnected = true
			}
			attempt++
			if !retryFn(ctx, attempt, err) {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
			continue
		}

		if err = ch.Switch(channel.ExpectAck, channel.PassErrors); err != nil {
			ch.Close()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !disconnected {
				if opts.disconnectHandler != nil {
					opts.disconnectHandler(err)
				}
				disconnected = true
			}
			attempt++
			if !retryFn(ctx, attempt, err) {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
			continue
		}

		// registration successful
		attempt = 0
		if disconnected {
			if opts.reconnectHandler != nil {
				opts.reconnectHandler()
			}
		}
		disconnected = false

		done := make(chan struct{})
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

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !disconnected {
			if opts.disconnectHandler != nil {
				opts.disconnectHandler(err)
			}
			disconnected = true
		}
	}
}

// newDefaultRetryHandler returns a RetryHandler that uses exponential backoff via sig.Retry.
func newDefaultRetryHandler() RetryHandler {
	retry, _ := sig.NewRetry(250*time.Millisecond, 10*time.Second, 2)
	return func(ctx *astral.Context, attempt int, err error) bool {
		if attempt == 1 {
			retry.Reset()
		}
		select {
		case <-retry.Retry():
			return true
		case <-ctx.Done():
			return false
		}
	}
}
