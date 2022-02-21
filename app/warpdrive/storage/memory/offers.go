package memory

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
)

var _ api.OfferStorage = OffersRepo{}

type OffersRepo struct {
	offers api.Offers
}

func Incoming(core api.Core) OffersRepo {
	return OffersRepo{core.Cache.Incoming}
}

func Outgoing(core api.Core) OffersRepo {
	return OffersRepo{core.Cache.Outgoing}
}

func (r OffersRepo) Save(offer api.Offer) {
	r.offers[offer.Id] = &offer
}

func (r OffersRepo) Get() api.Offers {
	return r.offers
}
