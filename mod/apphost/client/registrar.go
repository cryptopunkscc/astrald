package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// Registrar is the default implementation of apphost.Registrar.
// It blocks until the first registration, then automatically reconnects with node in the background in case of connection loss.
type Registrar struct {
	client *Client
}

var _ apphost.Registrar = &Registrar{}

func NewRegistrar(client *Client) *Registrar {
	return &Registrar{client: client}
}

func DefaultRegistrar() *Registrar {
	return NewRegistrar(Default())
}

func (r *Registrar) Register(ctx *astral.Context, endpoint string, token astral.Nonce) error {
	firstReg := make(chan error, 1)

	go func() {
		var registered bool

		for {
			bindChannel, err := r.client.Bind(ctx)
			if err != nil {
				if !registered {
					firstReg <- err
				}
				return
			}

			if err = r.client.RegisterHandler(ctx, endpoint, token); err != nil {
				bindChannel.Close()
				continue
			}

			if err = bindChannel.Send(&apphost.BindMsg{Token: token}); err != nil {
				bindChannel.Close()
				continue
			}

			if !registered {
				firstReg <- nil
				registered = true
			}

			for {
				if _, err = bindChannel.Receive(); err != nil {
					break
				}
			}
			bindChannel.Close()
		}
	}()

	return <-firstReg
}
