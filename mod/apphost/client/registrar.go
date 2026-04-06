package apphost

import (
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	defaultRetryMin    = 250 * time.Millisecond
	defaultRetryMax    = 30 * time.Second
	defaultRetryFactor = 2.0
)

type Registrar struct {
	client       *Client
	retry        *sig.Retry // nil = NoReconnect
	onConnect    func()
	onDisconnect func(error)
	onRetry      func(int, error)
	onFailed     func(int, error)
	ready        sig.Value[chan struct{}]
}

var _ apphost.Registrar = &Registrar{}

// RegistrarOption configures a Registrar at construction time.
type RegistrarOption func(*Registrar)

func WithRetry(r *sig.Retry) RegistrarOption {
	return func(reg *Registrar) { reg.retry = r }
}

func WithOnConnect(f func()) RegistrarOption {
	return func(r *Registrar) { r.onConnect = f }
}

func WithOnDisconnect(f func(error)) RegistrarOption {
	return func(r *Registrar) { r.onDisconnect = f }
}

func WithOnRetry(f func(attempt int, err error)) RegistrarOption {
	return func(r *Registrar) { r.onRetry = f }
}

func WithOnFailed(f func(attempt int, err error)) RegistrarOption {
	return func(r *Registrar) { r.onFailed = f }
}

func NewRegistrar(c *Client, opts ...RegistrarOption) *Registrar {
	r := &Registrar{
		client:       c,
		onConnect:    func() {},
		onDisconnect: func(error) {},
		onRetry:      func(int, error) {},
		onFailed:     func(int, error) {},
	}
	r.ready.Set(make(chan struct{}))
	r.retry, _ = sig.NewRetry(defaultRetryMin, defaultRetryMax, defaultRetryFactor)
	for _, o := range opts {
		o(r)
	}
	return r
}

func DefaultRegistrar() *Registrar {
	return NewRegistrar(Default())
}

func (r *Registrar) Ready() <-chan struct{} {
	return r.ready.Get()
}

func (r *Registrar) Register(ctx *astral.Context, endpoint string, token astral.Nonce) error {
	firstReg := make(chan error, 1)
	var once sync.Once
	signal := func(err error) { once.Do(func() { firstReg <- err }) }

	go func() {
		for {
			bindChannel, err := r.client.Bind(ctx)
			if err != nil {
				if r.retry == nil {
					r.onFailed(0, err)
					signal(err)
					return
				}
				select {
				case n := <-r.retry.Retry():
					r.onRetry(n, err)
				case <-ctx.Done():
					signal(ctx.Err())
					return
				}
				continue // retry Bind
			}
			if r.retry != nil {
				r.retry.Reset()
			}

			if err = r.client.RegisterHandler(ctx, endpoint, token); err != nil {
				bindChannel.Close()
				signal(err)
				return // protocol error — do not retry
			}

			if err = bindChannel.Send(&apphost.BindMsg{Token: token}); err != nil {
				bindChannel.Close()
				signal(err)
				return // protocol error — do not retry
			}

			close(r.ready.Get())
			signal(nil)
			r.onConnect()

			for {
				if _, err = bindChannel.Receive(); err != nil {
					break
				}
			}
			bindChannel.Close()

			r.ready.Set(make(chan struct{}))
			r.onDisconnect(err)
		}
	}()

	return <-firstReg
}
