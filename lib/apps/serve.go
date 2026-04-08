package apps

import (
	"github.com/cryptopunkscc/astrald/astral"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
)

// Serve creates a handler, registers it with the default session, gates queries until
// registered, and routes all inbound queries to router. Blocks until ctx is cancelled.
func Serve(ctx *astral.Context, router astral.Router) error {
	return ServeWith(ctx, router, NewDefaultAppRegistrar(ctx))
}

// ServeWith is like Serve but with an explicit registrar.
// If reg also implements libastrald.ReadyGate, queries are gated until it signals ready.
func ServeWith(ctx *astral.Context, router astral.Router, reg Registrar) error {
	h, err := NewHandler()
	if err != nil {
		return err
	}

	if rg, ok := reg.(libastrald.ReadyGate); ok {
		router = NewGateRouter(router, rg)
	}

	if err := reg.Register(ctx, h.Endpoint(), h.Token()); err != nil {
		h.Close()
		return err
	}

	return h.Route(ctx, router)
}
