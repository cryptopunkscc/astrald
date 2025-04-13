package user

import "github.com/cryptopunkscc/astrald/mod/status"

var _ status.Composer = &Module{}

func (mod *Module) ComposeStatus(a status.Composition) {
	if mod.config.Public {
		c := mod.ActiveContract()
		if c == nil {
			return
		}
		a.Attach(c)
	}
}
