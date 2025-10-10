package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/ip"
)

func (mod *Module) FindIPCandidates() (list []ip.IP) {
	if mod.lastObservedIP != nil {
		list = append(list, mod.lastObservedIP)
	}

	return
}
