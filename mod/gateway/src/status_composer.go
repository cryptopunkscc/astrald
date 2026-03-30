package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

var _ nearby.Composer = &Module{}

func (mod *Module) ComposeStatus(a nearby.Composition) {
	switch mod.Nearby.Mode() {
	case nearby.ModeSilent:
		// no-op
	case nearby.ModeVisible:
		for _, gw := range mod.gateways.Clone() {
			a.Attach(nodes.NewEndpointWithTTL(gateway.NewEndpoint(gw, mod.node.Identity()), 7*30*24*time.Hour))
		}
	case nearby.ModeStealth:
	}
}
