package setup

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/id"
)

func Context(ctx *handler.Context) {
	// API
	if ctx.Api == nil {
		ctx.Api = astral.Instance()
	}

	// Identity
	identity, err := id.Query()
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
