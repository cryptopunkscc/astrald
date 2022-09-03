package memory

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
)

var _ storage.Peer = Peers{}

type Peers core.Component

func (r Peers) Save(peers []warpdrive.Peer) {
	for _, peer := range peers {
		p := peer
		r.Cache.Peers[peer.Id] = &p
	}
}

func (r Peers) Get() warpdrive.Peers {
	return r.Cache.Peers
}

func (r Peers) List() (peers []warpdrive.Peer) {
	p := r.Get()
	for _, peer := range p {
		peers = append(peers, *peer)
	}
	return
}
