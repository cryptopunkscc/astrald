package tor

import (
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

var _ nearby.Composer = &Module{}

func (mod *Module) ComposeStatus(a nearby.Composition) {
	switch mod.Nearby.Mode() {
	case nearby.ModeSilent:
		// no-op
	case nearby.ModeVisible:
		if mod.torServer == nil || mod.torServer.endpoint.IsZero() {
			return
		}
		a.Attach(mod.torServer.endpoint)
	case nearby.ModeStealth:
	}
}
