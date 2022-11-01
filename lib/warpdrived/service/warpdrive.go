package service

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

var _ warpdrive.Service = Warpdrive{}

type Warpdrive core.Component

func (w Warpdrive) Incoming() warpdrive.OfferService {
	return incoming(core.Component(w))
}

func (w Warpdrive) Outgoing() warpdrive.OfferService {
	return outgoing(core.Component(w))
}

func (w Warpdrive) Peer() warpdrive.PeerService {
	return peer(w)
}

func (w Warpdrive) File() warpdrive.FileService {
	return w.FileResolver
}
