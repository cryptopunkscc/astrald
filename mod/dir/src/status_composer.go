package dir

import (
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

var _ nearby.Composer = &Module{}

func (mod *Module) ComposeStatus(a nearby.Composition) {
	switch mod.Nearby.Mode() {
	case nearby.ModeSilent:
		// no-op
	case nearby.ModeVisible:
		alias, err := mod.GetAlias(mod.node.Identity())
		if err == nil && alias != "" {
			v := dir.Alias(alias)
			a.Attach(&v)
		}
	case nearby.ModeStealth:
	}
}
