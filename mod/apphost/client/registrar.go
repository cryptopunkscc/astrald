// mod/apphost/client/registrar.go
package apphost

import (
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// Registrar manages the bind-session lifecycle with the host.
// Register blocks until the first registration succeeds, then returns.
// The background goroutine reconnects automatically per policy on session drop.
// Protocol errors (RegisterHandler or BindMsg failures) are fatal and stop retries.
type Registrar struct {
	client       *Client
	policy       astrald.RetryPolicy
	onConnect    func()
	onDisconnect func(error)
	onRetry      func(int, error)
	onFailed     func(int, error)

	mu    sync.RWMutex
	ready chan struct{}
}

var _ apphost.Registrar = &Registrar{}
var _ astrald.ReadyGate = &Registrar{}

// RegistrarOption configures a Registrar at construction time.
type RegistrarOption func(*Registrar)

func WithReconnectPolicy(p astrald.RetryPolicy) RegistrarOption {
	return func(r *Registrar) { r.policy = p }
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
		client: c,
		ready:  make(chan struct{}),
	}
	for _, o := range opts {
		o(r)
	}
	if r.policy == nil {
		r.policy = astrald.Backoff(250*time.Millisecond, 30*time.Second, 2.0)
	}
	return r
}

func DefaultRegistrar() *Registrar {
	return NewRegistrar(Default())
}

// Ready returns a channel that is closed when the bind session is established.
// It is replaced with a new open channel on session drop, then closed again on reconnect.
func (r *Registrar) Ready() <-chan struct{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ready
}

func (r *Registrar) setReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.ready)
}

func (r *Registrar) setNotReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ready = make(chan struct{})
}

func (r *Registrar) Register(ctx *astral.Context, endpoint string, token astral.Nonce) error {
	firstReg := make(chan error, 1)

	go func() {
		var registered bool

		for attempt := 0; ; {
			bindChannel, err := r.client.Bind(ctx)
			if err != nil {
				d, ok := r.policy.Next(attempt, err)
				if !ok {
					if r.onFailed != nil {
						r.onFailed(attempt, err)
					}
					if !registered {
						firstReg <- err
					}
					return
				}
				if r.onRetry != nil {
					r.onRetry(attempt, err)
				}
				attempt++
				select {
				case <-time.After(d):
				case <-ctx.Done():
					if !registered {
						firstReg <- ctx.Err()
					}
					return
				}
				continue
			}
			attempt = 0 // Bind succeeded — reset consecutive-failure counter

			if err = r.client.RegisterHandler(ctx, endpoint, token); err != nil {
				bindChannel.Close()
				if !registered {
					firstReg <- err
				}
				return // protocol error — do not retry
			}

			if err = bindChannel.Send(&apphost.BindMsg{Token: token}); err != nil {
				bindChannel.Close()
				if !registered {
					firstReg <- err
				}
				return // protocol error — do not retry
			}

			r.setReady()
			if !registered {
				firstReg <- nil
				registered = true
			}
			if r.onConnect != nil {
				r.onConnect()
			}

			for {
				if _, err = bindChannel.Receive(); err != nil {
					break
				}
			}
			bindChannel.Close()

			r.setNotReady()
			if r.onDisconnect != nil {
				r.onDisconnect(err)
			}
		}
	}()

	return <-firstReg
}
