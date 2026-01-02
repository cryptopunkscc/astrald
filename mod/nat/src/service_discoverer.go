package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
)

func (mod *Module) DiscoverService(ctx *astral.Context, caller *astral.Identity) (*services.Service, <-chan services.ServiceChange, error) {
	snapshot := mod.serviceFeed.GetSnapshot()
	ch := mod.serviceFeed.Subscribe(ctx)

	// Possibility of filtering based on caller's identity can be added here

	return snapshot, ch, nil
}
