package core

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"log"
)

func New(config service.Config) api.Core {
	return &core{Config: config}
}

type core struct {
	service.Config
	*log.Logger
	*persistence
	*cache
	*observers
}

type cache struct {
	incoming api.Offers
	outgoing api.Offers
	peers    api.Peers
}

type persistence struct {
	api.Repository
	api.Resolver
	api.Storage
}

type observers struct {
	filesOffers    *api.Subscriptions
	incomingStatus *api.Subscriptions
	outgoingStatus *api.Subscriptions
}
