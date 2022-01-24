package core

import "github.com/cryptopunkscc/astrald/app/warpdrive/api"

type offerManager struct {
	*core
	*api.Subscriptions
	repo   api.OffersRepo
	offers api.Offers
}

func (c *core) Outgoing() api.OfferManager {
	return &offerManager{
		core:          c,
		Subscriptions: c.outgoingStatus,
		repo:          c.Repository.Outgoing(),
		offers:        c.outgoing,
	}
}

func (c *core) Incoming() api.OfferManager {
	return &offerManager{
		core:          c,
		Subscriptions: c.incomingStatus,
		repo:          c.Repository.Incoming(),
		offers:        c.incoming,
	}
}

func (c *offerManager) Add(offerId string, files []api.Info, peer *api.Peer) {
	offer := &api.Offer{
		Files: files,
		Status: api.Status{
			Id: api.OfferId(offerId),
		},
	}
	if peer == nil {
		offer.Status.Status = "sent"
	} else {
		offer.Status.Status = "received"
		offer.Peer = peer.Id
	}

	c.offers[offer.Id] = offer
	c.repo.Save(offer)

	go c.notifyListeners(offer.Status, c.Subscriptions)

	if peer != nil && peer.Mod == api.PeerModAsk {
		go c.notifyListeners(offer, c.filesOffers)
	}
}

func (c *offerManager) Update(offer *api.Offer, status string, persist bool) {
	offer.Status.Status = status
	if persist {
		c.repo.Save(offer)
	}
	go c.notifyListeners(offer.Status, c.Subscriptions)
}

func (c *offerManager) Get(id api.OfferId) *api.Offer {
	return c.offers[id]
}

func (c *offerManager) List() api.Offers {
	return c.offers
}

func (c *offerManager) Status() *api.Subscriptions {
	return c.Subscriptions
}
