package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

const ViaRouterHintKey = "via"

type Router interface {
	net.Router
	AddRoute(caller id.Identity, target id.Identity, router net.Router, priority int) error
	RemoveRoute(caller id.Identity, target id.Identity, router net.Router) error
	Routes() []Route
}
