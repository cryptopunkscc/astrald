package apphost

import "github.com/cryptopunkscc/astrald/astral"

// Registrar manages the registration lifecycle of an app handler with the host.
// Register blocks until the endpoint is first registered, then returns.
// onClose is called when the Registrar permanently stops managing registration.
// The implementation controls whether and how to reconnect after a disconnect.
type Registrar interface {
	Register(ctx *astral.Context, endpoint string, token astral.Nonce, onClose func()) error
}
