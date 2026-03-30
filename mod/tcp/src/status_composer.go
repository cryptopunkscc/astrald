package tcp

import (
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

var _ nearby.Composer = &Module{}

func (mod *Module) ComposeStatus(a nearby.Composition) {
	switch mod.Nearby.Mode() {
	case nearby.ModeSilent:
		// no-op
	case nearby.ModeVisible:
		endpoints := mod.endpoints()
		for _, endpoint := range endpoints {
			a.Attach(endpoint)
		}
	case nearby.ModeStealth:
	}
}
