package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type API interface {
	SetRouter(target id.Identity, router id.Identity)
	GetRouter(target id.Identity) id.Identity
	Reroute(nonce net.Nonce, router net.Router) error
}
