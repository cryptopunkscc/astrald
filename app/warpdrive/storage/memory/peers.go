package memory

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
)

var _ api.PeerStorage = Peers{}

type Peers api.Core

func (r Peers) Save(peers []api.Peer) {
	for _, peer := range peers {
		p := peer
		r.Cache.Peers[peer.Id] = &p
	}
}

func (r Peers) Get() api.Peers {
	return r.Cache.Peers
}

func (r Peers) List() (peers []api.Peer) {
	p := r.Get()
	for _, peer := range p {
		peers = append(peers, *peer)
	}
	return
}
