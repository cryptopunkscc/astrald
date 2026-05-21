package user

import (
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) FindObject(ctx *astral.Context, id *astral.ObjectID) (<-chan *astral.Identity, error) {
	out := make(chan *astral.Identity)
	go func() {
		defer close(out)

		for _, sib := range mod.getLinkedSibs() {
			select {
			case out <- sib:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
