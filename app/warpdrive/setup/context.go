package setup

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func Context(ctx *handler.Context) {
	// API
	if ctx.Api == nil {
		ctx.Api = astral.AppHostAdapter()
	}

	// Identity
	identity, err := ctx.Api.Resolve("localnode")
	if err != nil {
		ctx.Panic("Cannot obtain node identity", err)
	}
	ctx.Identity = identity.String()

	// Peers
	service.Peer(ctx.Core).Fetch()

	// Offer updates
	stop := service.OfferUpdates(ctx.Core).Start()
	go func() {
		<-ctx.Done()
		stop()
	}()
}
