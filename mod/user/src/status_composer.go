package user

import "github.com/cryptopunkscc/astrald/mod/status"

var _ status.Composer = &Module{}

func (mod *Module) ComposeStatus(a status.Composition) {
	if mod.config.Public {
		c, err := mod.LocalContract()
		if err != nil || c == nil {
			return
		}
		a.Attach(c)
	}
}
