package apps

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	apphostclient "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/sig"
)

// Registrar registers an IPC endpoint with the node.
type Registrar interface {
	Register(ctx *astral.Context, endpoint string, token astral.Nonce) error
}

// AppRegistrarEvents are optional lifecycle callbacks.
// NodeBinding was an alternative name considered for AppRegistrar.
type AppRegistrarEvents struct {
	OnConnect    func()
	OnDisconnect func(error)
}

type handler struct {
	endpoint string
	token    astral.Nonce
}

type regRequest struct {
	handler
	done chan error
}

// AppRegistrar maintains a persistent bind channel to the node, re-registers all
// handlers on reconnect, and signals readiness. Implements Registrar and ReadyGate.
type AppRegistrar struct {
	client   *apphostclient.Client
	dial     NodeBind
	events   AppRegistrarEvents
	ready    sig.Value[chan struct{}]
	handlers []handler
	regCh    chan *regRequest
}

var _ Registrar = &AppRegistrar{}
var _ libastrald.ReadyGate = &AppRegistrar{}

type AppRegistrarOption func(*AppRegistrar)

// WithClient sets the apphost client and derives the node bind from it.
func WithClient(c *apphostclient.Client) AppRegistrarOption {
	return func(s *AppRegistrar) {
		s.client = c
		s.dial = c.Bind
	}
}

// WithBind overrides the node bind function.
func WithBind(d NodeBind) AppRegistrarOption { return func(s *AppRegistrar) { s.dial = d } }

// WithEvents sets lifecycle callbacks.
func WithEvents(e AppRegistrarEvents) AppRegistrarOption {
	return func(s *AppRegistrar) { s.events = e }
}

// NewAppRegistrar creates an AppRegistrar and starts its run loop with the given ctx.
func NewAppRegistrar(ctx *astral.Context, opts ...AppRegistrarOption) *AppRegistrar {
	c := apphostclient.Default()
	s := &AppRegistrar{
		client: c,
		dial:   NodeBind(c.Bind),
		regCh:  make(chan *regRequest),
	}
	s.ready.Set(make(chan struct{}))
	for _, o := range opts {
		o(s)
	}
	go s.run(ctx)
	return s
}

// NewDefaultAppRegistrar creates an AppRegistrar with default options and starts its run loop.
func NewDefaultAppRegistrar(ctx *astral.Context, opts ...AppRegistrarOption) *AppRegistrar {
	return NewAppRegistrar(ctx, opts...)
}

func (s *AppRegistrar) Ready() <-chan struct{} { return s.ready.Get() }

func (s *AppRegistrar) Register(ctx *astral.Context, endpoint string, token astral.Nonce) error {
	req := &regRequest{handler{endpoint, token}, make(chan error, 1)}
	if err := sig.Send(ctx, s.regCh, req); err != nil {
		return err
	}
	return sig.RecvErr(ctx, req.done)
}

func (s *AppRegistrar) run(ctx *astral.Context) {
	for {
		bindCh, err := s.dial(ctx)
		if err != nil {
			return
		}

		for _, h := range s.handlers {
			if err = s.attach(ctx, bindCh, h); err != nil {
				bindCh.Close()
				return
			}
		}

		close(s.ready.Get())
		if s.events.OnConnect != nil {
			s.events.OnConnect()
		}

		reconnect, disconnectErr := s.loop(ctx, bindCh)
		bindCh.Close()

		if !reconnect {
			return
		}

		s.ready.Set(make(chan struct{}))
		if s.events.OnDisconnect != nil {
			s.events.OnDisconnect(disconnectErr)
		}
	}
}

func (s *AppRegistrar) attach(ctx *astral.Context, bindCh *channel.Channel, h handler) error {
	if err := s.client.RegisterHandler(ctx, h.endpoint, h.token); err != nil {
		return err
	}
	return bindCh.Send(&apphost.BindMsg{Token: h.token})
}

func (s *AppRegistrar) loop(ctx *astral.Context, bindCh *channel.Channel) (reconnect bool, disconnectErr error) {
	disconnected := make(chan error, 1)
	go func() {
		for {
			if _, err := bindCh.Receive(); err != nil {
				disconnected <- err
				return
			}
		}
	}()

	for {
		select {
		case req := <-s.regCh:
			if err := s.attach(ctx, bindCh, req.handler); err != nil {
				req.done <- err
				return false, nil
			}
			s.handlers = append(s.handlers, req.handler)
			req.done <- nil

		case disconnectErr = <-disconnected:
			return true, disconnectErr

		case <-ctx.Done():
			return false, nil
		}
	}
}
