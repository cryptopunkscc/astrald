package objects

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) AddHolder(h objects.Holder) error {
	return mod.holders.Add(h)
}

func (mod *Module) Holders(objectID *object.ID) (holders []objects.Holder) {
	for _, h := range mod.holders.Clone() {
		if h.HoldObject(objectID) {
			holders = append(holders, h)
		}
	}
	return
}
