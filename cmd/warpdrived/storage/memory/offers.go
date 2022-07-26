package memory

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

func Incoming(core core.Component) storage.Offer {
	return offer(core.Cache.Incoming)
}

func Outgoing(core core.Component) storage.Offer {
	return offer(core.Cache.Outgoing)
}

type offer warpdrive.Offers

func (r offer) Save(offer warpdrive.Offer) {
	r[offer.Id] = &offer
}

func (r offer) Get() warpdrive.Offers {
	return warpdrive.Offers(r)
}
