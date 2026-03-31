package apphost

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// Registrar is the default implementation of apphost.Registrar.
// It blocks until first registration, then automatically reconnects in the background.
// onClose is never called — the Registrar reconnects indefinitely until ctx is done.
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

func (r *Registrar) Register(ctx *astral.Context, endpoint string, token astral.Nonce, onClose func()) error {
	firstReg := make(chan error, 1)

	go func() {
		var once sync.Once

		for {
			bc, err := r.client.Bind(ctx)
			if err != nil {
				once.Do(func() { firstReg <- err })
				return
			}

			if err = r.client.RegisterHandler(ctx, endpoint, token); err != nil {
				bc.Close()
				continue
			}

			if err = bc.Send(&apphost.BindMsg{Token: token}); err != nil {
				bc.Close()
				continue
			}

			once.Do(func() { firstReg <- nil })

			for {
				if _, err = bc.Receive(); err != nil {
					break
				}
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return <-firstReg
}
